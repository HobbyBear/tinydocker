# RepoMind — 代码问答与编码的优先知识库

## 核心原则

- 任何涉及代码、业务逻辑、项目结构、异常排查，需求分析，方案设计的问题，都必须先查 RepoMind，再回答或改代码。
- RepoMind 查出来的内容不是“参考一下就算了”，而是回答结论、修改决策、排查路径的凭证和上下文依据。
- 命中的 concepts / modules / troubles 以及必要的 graphify 结构结果，必须真正进入回答或实现判断；不能查完不用，也不能绕开检索结果直接下结论。
- 如果 RepoMind 命中结果不足以支持结论，必须明确说“当前证据不足”，并继续补查代码或图谱。
- RepoMind 当前采用每个 knowledge 文档 frontmatter 里的 `name` / `description` 元数据做首轮路由；其中 `description` 是首要索引摘要，模块文档还要额外维护 `keywords` 作为辅助定位词，不依赖集中式 `index.json` 或目录 README。

## repomind-query 触发时机

以下场景必须先触发 `repomind-query`：

1. 用户询问业务概念、业务规则、项目结构、代码定位、异常现象时。
2. 准备编辑或修改业务代码前。
3. 排查 Bug、分析“为什么没生效 / 为什么不对 / 是不是 Bug”时。
4. 处理历史 PRD 前，如果需要先理解现有业务知识和模块上下文。
5. 用户纠正 AI 或 RepoMind 的业务结论、模块判断或排查结论时，例如“X 才是对的”“Y 错了”“不是 A，是 B”。

## repomind-query 使用要求

1. 先查知识库元数据，再按命中打开 concepts / modules / troubles；模块路由要同时参考 `name` / `keywords` / `description`。需要代码证据时先用模块文档入口配合平台代码搜索工具取小上下文：Claude Code 用 Grep 定位后 Read 最小片段，Codex/终端用 `rg -n -C 3`。上下文足够回答或修改时停止扩展读取。只有需要调用链、影响面或跨模块关系时，才补查 `graphify query` / `explain` / `path`。`.repomind/graph/summary.json` 只作为初始化辅助摘要，不作为 query 阶段默认检索入口；普通 query 不读取 `graphify-out/graph.json` 或 `GRAPH_REPORT.md`。
2. 最终回答必须基于命中的知识组织，而不是把检索结果放在一边。
3. 如果命中了业务卡片，回答里要体现业务定义、边界或预期。
4. 如果命中了模块文档，回答或改动方案里要体现关键入口、影响范围或注意事项。
5. 如果命中了排查记录，回答里要体现历史现象、判断顺序或常见根因。
6. 如果命中内容和当前代码冲突，以当前代码为准，并明确指出冲突。
7. 需要调用链、调用方或被调用方时，优先使用 graphify 的结构化结果；AI 不能凭少量源码片段声称完整 callers / callees。
8. 如果本轮代码定位不是直接通过现有模块文档完成，而是绕过模块文档去查代码搜索 / graphify / source 才定位到实现，那么本轮结束前必须触发 `repomind-summary`，把缺失的入口信息或模块关键词补回 RepoMind。
9. 只要用户给出业务纠错或修订结论，就必须把纠错内容写入 `.repomind/.query-findings.json`，并令 `needs_summary = true`。
10. 每次执行过 `repomind-query` 后，最终答复前都必须进入一次 `repomind-summary` 的 summary gate；即使 gate 最终判定无需更新，也不能跳过 gate。
11. 每次完成代码修改、生成文件、修复 bug 或跑完验证后，最终答复前也必须进入一次 `repomind-summary` 的 summary gate；不能因为“只是写代码”就跳过 gate。
12. `repomind-summary` 是同步阻塞步骤：不要说“summary 正在运行”就继续回答；必须等它真正完成。如果当前平台不能显式嵌套调用 skill，就在当前流程里直接执行 summary 步骤。

## repomind-summary 触发时机

以下场景必须触发 `repomind-summary`：

1. 每次 `repomind-query` 完成后，最终答复前必须触发一次 summary gate；gate 可判定无需更新，但 gate 本身不能省略。
2. 每次代码修改、生成文件、修复 bug 或跑完验证后，最终答复前必须触发一次 summary gate；gate 可判定无需更新，但 gate 本身不能省略。
3. 问答完成后，只要形成了可复用的新业务知识、模块知识或排查经验。
4. 用户纠正业务事实、模块归属、入口位置、排查根因或历史结论时，例如“X 才是”“Y 错了”“不是 A，是 B”。
5. 用户明确要求沉淀知识时，例如“记一下”“总结到知识库”“以后遇到这个要注意”“这个经验要沉淀”。
6. 业务讨论、需求分析、PRD 同步后，只要确认了新的概念边界、规则、历史原因或业务意图。
7. 排查结束后，只要形成了可复用的现象、判断路径、根因、验证方式或修订结论。
8. 本轮存在绕过现有模块文档、依赖代码搜索 / graphify / source 才完成代码定位时，即使最后只补入口或关键词，也必须触发。
9. 本轮识别出某个模块应新增、删除或收紧 `keywords` 时，也必须触发。

## repomind-summary 使用要求

1. 先做 summary gate，再决定是否落库。
2. 只沉淀代码不容易直接看出的知识，不重复写显式源码细节。
3. summary 时先维护索引元数据，再维护正文；优先检查 `description` 是否还适合作为首轮路由摘要，模块文档还要同步检查 `keywords` 是否覆盖最新别称、入口词和常见搜索词。
4. 发现新知识后不要拖到以后；本轮结束前就闭环到 RepoMind。
5. 如果本轮通过直接代码查找才找到答案，至少要把“缺失的模块入口 / 新增关键词 / 应补的常见修改场景”总结回 RepoMind。
6. 如果本轮是用户纠错，必须把旧说法、新说法、证据来源和影响范围写入对应 concept/module/trouble；不能只在对话里口头承认。
7. 如果本轮是用户手动要求沉淀，必须判断它更像业务概念、模块修改经验还是排查经验，只写入 concepts/modules/troubles，不创建新的集中式导览或索引文档。
8. 代码写完后的 summary gate 也必须同步完成；即使最终不写库，也要完成 gate 判定后再给用户最终答复。
9. 不允许把 summary 描述成后台任务；只有在 summary 完成或明确判定无需更新之后，才能给用户最终答复。