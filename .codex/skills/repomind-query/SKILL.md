---
name: repomind-query
description: 查阅业务逻辑、定位代码、排查问题，需求分析，方案设计时优先自动触发。先用每个 knowledge 文档的 name/description 元数据做 skill-style 路由，再按需打开 concepts、modules、troubles 和最小代码证据；代码定位优先使用模块文档入口和平台代码搜索小上下文，只有调用链、影响面或跨模块关系不足时才补查 graphify query/explain/path，回答前自动进入 repomind-summary gate；有新发现或用户纠错时写回 RepoMind。
metadata:
  short-description: 先查 RepoMind 再回答
---

# RepoMind 编码前 / 问答分析

任何涉及业务逻辑、代码修改、代码定位、项目结构、异常排查的提问，都必须先执行本流程。

纯技术问题可以跳过，例如依赖安装、语言语法、编译器通用报错。

## 核心原则

1. 先识别意图维度，再决定查哪些知识源。
2. 先读元数据，再决定打开哪些正文。
3. 路由不依赖 `index.json` 或 README；优先依赖各知识文档自己的 `name` / `description`，其中 `modules` 还要额外使用 `keywords`。
4. 只把“代码不会直接告诉你的新知识”写入 `.repomind/.query-findings.json`。
5. 每次执行本流程后，最终答复前都必须进入 `repomind-summary` 的 summary gate；gate 可以判定无需更新，但不能省略。
6. 用户纠正业务事实、模块归属、入口位置、排查根因或历史结论时，必须视为新发现并令 `needs_summary = true`。

## 知识源边界

### concepts

负责回答：

- 某个业务概念是什么
- 为什么有
- 用户侧表现
- 业务边界和易混淆概念

frontmatter `description` 必须能回答：

- 这个概念会在哪些语境下被提起
- 命中后为什么值得继续打开这张卡

### modules

负责回答：

- 关键入口在哪
- 要改哪些模块
- 影响面和隐性约束是什么

frontmatter `description` 必须能回答：

- 这个模块管什么业务
- 什么场景需要打开它
- 典型跨模块风险是什么

frontmatter `keywords` 必须承担：

- 模块名和常见别称
- 英文名、缩写、核心业务词
- 用户最可能拿来搜这个模块的 3-8 个判别词

### troubles

负责回答：

- 类似问题以前怎么排查
- 常见根因和验证路径是什么

frontmatter `description` 必须能回答：

- 这类问题的典型现象是什么
- 首查方向是什么

### 代码搜索 / graphify / 代码证据

负责提供当前代码事实，不承担业务解释，也不让 AI 自行推断完整调用图：

- 文件/函数定义
- 代码搜索小上下文中的参数、返回值、分支条件、副作用
- graphify 输出的结构关系，例如文件、函数、模块、社区、路径和显式边

查询阶段的主入口是 RepoMind 文档；平台代码搜索工具是代码定位主力，graphify 是关系补充，不是业务解释入口：

- Claude Code 中使用 Grep 定位文件/行，再用 Read 读取命中行附近的最小片段；Codex/终端中使用 `rg -n -C 3`。
- 模块文档命中并且代码搜索小上下文已经足以回答或修改时，不要继续大量读源码；否则模块文档的入口映射就失去意义。
- 只有需要调用链、影响面、跨模块关系，或代码搜索只能找到零散候选但无法判断关系时，才补查 graphify。
- graphify 查询必须带预算，优先小预算定位，避免把图谱结果当成全文背景塞进上下文。
- 不得把少量源码片段中看到的调用关系说成完整 callers / callees；只有 graphify 明确输出，或当前打开文件中直接出现的调用，才能作为调用证据，并标明是否非穷尽。
- `.repomind/graph/summary.json` 是初始化阶段的模块候选摘要，不作为 query 阶段的默认检索入口；除非当前流程已经打开了它，否则不要为回答用户问题专门读取它。

## 步骤 0：先修正旧格式

在任何查询前，先执行：

```bash
repomind kb-migrate
```

如果仓库里还残留旧格式，这一步会先修复，再继续查询。

## 步骤 1：读取知识库元数据

先执行：

```bash
repomind kb-metadata
```

读取 JSON 中三类信息：

- `concepts[].name/description`
- `modules[].name/keywords/description`
- `troubles[].name/description`

这一阶段只做 skill-style 首轮匹配，不打开全部 markdown 正文。

## 步骤 2：识别意图维度

用户的问题可能同时包含多个维度：

- 业务概念：是什么、为什么、边界、区别、预期
- 代码模块：在哪改、入口在哪、影响哪些模块
- 异常排查：为什么没生效、是不是 bug、怎么排查
- 业务纠错：用户指出“X 才是”“Y 错了”“不是 A，是 B”，或推翻 AI / RepoMind 的旧结论

回答前先做内部判断：

| 维度 | 是否激活 | 依据 |
|------|----------|------|
| 业务概念 | ✅/❌ | 命中的业务词、预期、边界问题 |
| 代码模块 | ✅/❌ | 改哪里、入口、调用链、影响范围 |
| 异常排查 | ✅/❌ | 没生效、数据不一致、报错、异常 |
| 业务纠错 | ✅/❌ | 用户明确修正事实、边界、模块归属、入口、根因或历史结论 |

## 步骤 3：先用元数据选文档，再打开正文

### 3a：concepts 路由

当业务概念维度激活时：

1. 先用问题里的业务词与 `concepts[].name/description` 做语义匹配。
2. 只打开最相关的 1-3 张 concept 卡片。
3. 从正文提炼定义、预期、边界、易混淆概念。

### 3b：modules 路由

当代码模块维度激活时：

1. 先用问题里的业务域、改动意图、接口/模块名与 `modules[].name/keywords/description` 匹配。
2. 只打开最相关的 1-3 份模块文档。
3. 从正文提炼关键入口、修改场景、AI 注意事项。
4. 如果模块文档已给出具体入口，先用平台代码搜索工具对入口名、函数名、接口名或业务关键词做小上下文验证；上下文足够时停止扩展读取。

### 3c：troubles 路由

当异常排查维度激活时：

1. 先用症状、影响面、异常词与 `troubles[].name/description` 匹配。
2. 只打开命中的排查记录。
3. 提炼现象、判断顺序、根因、验证方式。

### 3d：补查代码证据

只有以下场景才补查代码证据、图谱或结构化查询：

- 元数据命中了，但正文不够精确
- 正文命中了，但需要用小上下文确认入口、签名、关键分支或副作用是否仍成立
- 三类知识源互相冲突
- 用户明确要调用链、函数位置、影响范围

优先顺序：

1. 命中的模块文档给出的入口
2. 平台代码搜索工具精确搜索入口名 / 函数名 / 接口名 / 业务关键词，并只保留少量上下文
3. 需要调用链、影响路径或跨模块关系时，再用 `graphify query` / `graphify explain` / `graphify path`
4. 仍不足时，读取最小源码片段

代码搜索方法：

1. 先从模块文档入口、函数名、接口名、业务关键词中选 1-3 个最强关键词。
2. Claude Code：使用 Grep 定位匹配文件/行，再用 Read 只读取命中行附近的最小片段；不要为了看上下文直接全文件 Read。
3. Codex/终端：使用 `rg -n -C 3` 获取小上下文：

   ```bash
   rg -n -C 3 "<入口名|函数名|接口名|业务关键词>"
   ```

4. 如果命中太多，先加路径范围或更精确关键词，不要扩大到全量阅读：

   ```bash
   rg -n -C 3 "<更精确关键词>" <模块文档给出的目录或文件>
   ```

5. 代码搜索小上下文已经能确认入口、签名、关键分支、副作用或修改点时，停止读取更多源码。
6. 只有小上下文不足以判断局部行为时，才打开对应文件的最小片段。

graphify 查询方法：

1. 只有需要调用链、影响面、跨模块关系，或代码搜索定位不到可靠入口时，才做小预算定位：

   ```bash
   graphify query "定位与 <业务词/接口/函数/模块> 相关的代码入口、调用关系或影响路径，返回最可能的文件、函数、模块和直接依据，不要展开无关背景。" --budget 800
   ```

2. 如果 query 命中具体节点，再解释该节点：

   ```bash
   graphify explain "<命中的文件/函数/模块/节点名>"
   ```

3. 如果用户要求 A 到 B 的关系、调用链或影响路径，使用 path：

   ```bash
   graphify path "<入口函数/模块/文件>" "<目标函数/模块/文件>"
   ```

4. 如果 graphify CLI 不可用，不要改读 `graphify-out/graph.json` 或 `GRAPH_REPORT.md`；继续用平台代码搜索和最小源码片段兜底。
5. 如果 graphify 输出只给出候选文件，没有给出明确调用边，则只能把它当作定位线索，再读取候选入口的最小源码片段确认当前事实。

调用关系约束：

- AI 不负责从源码片段自行归纳完整调用图。
- 只有 graphify 显式边、`graphify path` 结果，或当前打开文件中的直接调用，才能作为 callers / callees 证据。
- 仅从当前文件看到的调用必须标为“局部直接证据”，不得说成全量调用链。
- interface 动态分发、反射、依赖注入、路由注册、函数变量等场景必须标明证据边界；工具没有给出高置信结果时，说“调用链证据不足”。
- 禁止为了确认而全量扫源码。

**强制总结规则：**

- 如果本轮代码定位不是直接从现有 `modules` 文档命中入口，而是依赖代码搜索、graphify 或源码片段继续找出来的，那么本轮结束前必须触发一次 `repomind-summary`
- 原因不是“查了代码就一定有新知识”，而是这通常意味着现有模块文档或模块关键词不足以完成首轮路由
- 此时至少要写一条 `module_knowledge`，说明：
  - 缺了哪个入口定位
  - 缺了哪些模块关键词/别称/入口词
  - 以后用户再问同类问题时，模块文档应该如何更快命中

## 步骤 4：根据维度组合查询顺序

| 激活维度 | 推荐顺序 |
|----------|----------|
| 仅业务概念 | concepts → 必要时 modules/code |
| 仅代码模块 | modules → 代码搜索小上下文 → 必要时 graphify → 最小源码片段 |
| 仅异常排查 | troubles → concepts → modules → 代码搜索小上下文 → 必要时 graphify → 最小源码片段 |
| 业务概念 + 代码模块 | concepts → modules → 代码搜索小上下文 → 必要时 graphify → 最小源码片段 |
| 业务概念 + 异常排查 | concepts → troubles → modules |
| 代码模块 + 异常排查 | troubles → modules → 代码搜索小上下文 → 必要时 graphify → 最小源码片段 |
| 三者都有 | concepts → troubles → modules → 代码搜索小上下文 → 必要时 graphify → 最小源码片段 |
| 业务纠错 | 先定位被纠错的 concept/module/trouble → 当前代码或用户确认补证 → 保存修订发现 |

不要为了覆盖率把整个知识库都读一遍。

## 步骤 5：形成回答依据

回答前必须明确：

- 哪些文档命中后被真正用作结论依据
- 哪些只是背景参考
- 是否与当前代码或其他知识源冲突
- 是否还缺证据

回答要求：

- 业务结论必须能追溯到 concept 卡片、用户确认或当前代码。
- 代码位置必须能追溯到 module 文档和当前代码证据。
- 调用方/被调用方必须能追溯到 graphify 结构结果，或明确打开文件中的局部直接调用；证据不全时不得声称完整调用链。
- 排查建议必须能追溯到 trouble 记录、当前代码或本次分析形成的明确路径。
- 证据不足时必须直说“当前证据不足”，并说明下一步查什么。

## 步骤 6：保存新发现

如果本次查询发现了超出已有知识库的新知识，写入 `.repomind/.query-findings.json`。

只记录三类：

- `concept_knowledge`
- `module_knowledge`
- `trouble_knowledge`

兼容旧类型时：

- `index_knowledge` 视为 `module_knowledge`
- `new_business_card` / `new_business_rule` 视为 `concept_knowledge`
- `module_update` / `new_code_location` 视为 `module_knowledge`
- `trouble_record` 视为 `trouble_knowledge`

模板：

```bash
cat > .repomind/.query-findings.json << 'JSONEOF'
{
  "trigger": "问答",
  "intent": "用户意图简述",
  "known_modules": ["已命中模块"],
  "new_findings": [
    {
      "type": "concept_knowledge|module_knowledge|trouble_knowledge",
      "module": "主模块名",
      "file": "concepts/xxx.md 或 modules/xxx.md",
      "content": "新发现描述"
    }
  ],
  "needs_summary": true
}
JSONEOF
```

规则：

- `new_findings` 为空时，`needs_summary` 必须为 `false`
- 只要存在任何新知识，`needs_summary` 就必须为 `true`
- 如果本轮发生了“绕过 module 文档、依赖代码搜索 / graphify / 源码片段才定位到实现”的情况，即使最后只补模块入口词或关键词，也必须令 `needs_summary = true`
- 如果用户纠正了 AI 或 RepoMind 的业务事实、模块归属、入口位置、排查根因或历史结论，必须记录旧说法、新说法、证据来源和影响范围，并令 `needs_summary = true`

## 步骤 7：自动触发 summary

每次执行本流程后，最终答复前都必须自动调用：

```text
Skill: repomind-summary
```

这是一个**同步阻塞步骤**，不是后台任务。

- 不要输出“summary skill is running”“let me wait for it”“它会自己处理”之类的话
- 不要在 `repomind-summary` 完成前就回到主回答
- `needs_summary == true` 时，summary 必须处理 `.repomind/.query-findings.json`
- `needs_summary == false` 或没有新发现时，也必须进入 summary gate，并等它明确判定“无需更新”后，才能给用户最终回复
- 如果当前平台无法在 skill 内再次显式调用 skill，就在当前流程里直接执行 `repomind-summary` 的完整步骤，而不是口头移交

## 步骤 8：持续发现

对话过程中如果又出现了新的业务语义、模块边界或排查经验：

1. 直接追加到 `.repomind/.query-findings.json`
2. 调用 `repomind-summary`

不需要重新跑 query。

如果持续发现的是：

- 模块文档缺入口
- 模块关键词不够
- 用户常用叫法没进 `keywords`
- 用户纠正了旧业务结论、模块归属或排查根因
- 用户明确要求“记一下 / 总结到知识库 / 以后遇到这个要注意 / 这个经验要沉淀”

也同样要按类型写入 `concept_knowledge` / `module_knowledge` / `trouble_knowledge`，然后触发 `repomind-summary`。
