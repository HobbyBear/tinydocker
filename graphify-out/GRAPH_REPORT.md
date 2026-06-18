# Graph Report - .  (2026-06-18)

## Corpus Check
- Corpus is ~9,183 words - fits in a single context window. You may not need a graph.

## Summary
- 131 nodes · 231 edges · 12 communities (9 shown, 3 thin omitted)
- Extraction: 91% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 18 edges (avg confidence: 0.83)
- Token cost: 93,197 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Container Lifecycle|Container Lifecycle]]
- [[_COMMUNITY_RepoMind Skills|RepoMind Skills]]
- [[_COMMUNITY_Bridge Network Driver|Bridge Network Driver]]
- [[_COMMUNITY_Logging System|Logging System]]
- [[_COMMUNITY_IPAM Filesystem|IPAM Filesystem]]
- [[_COMMUNITY_Network Tests|Network Tests]]
- [[_COMMUNITY_Workspace Namespaces|Workspace Namespaces]]
- [[_COMMUNITY_Network Manager Config|Network Manager Config]]
- [[_COMMUNITY_RepoMind Rules|RepoMind Rules]]
- [[_COMMUNITY_Project Overview|Project Overview]]
- [[_COMMUNITY_Init Knowledge Strategy|Init Knowledge Strategy]]
- [[_COMMUNITY_KB Format Migration|KB Format Migration]]

## God Nodes (most connected - your core abstractions)
1. `bridgeDriver` - 15 edges
2. `ipAmFs` - 15 edges
3. `bitMap` - 11 edges
4. `repomind-summary Skill` - 11 edges
5. `Logger` - 10 edges
6. `main` - 10 edges
7. `main()` - 9 edges
8. `repomind-query Skill` - 9 edges
9. `SetMntNamespace()` - 8 edges
10. `delMntNamespace()` - 8 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `delMntNamespace()`  [INFERRED]
  main.go → workspace/workspace.go
- `main()` --calls--> `SetMntNamespace()`  [INFERRED]
  main.go → workspace/workspace.go
- `repomind-prd Skill` --semantically_similar_to--> `repomind-prd Skill (Codex)`  [EXTRACTED] [semantically similar]
  .claude/skills/repomind-prd/SKILL.md → .codex/skills/repomind-prd/SKILL.md
- `repomind-query Skill` --semantically_similar_to--> `repomind-query Skill (Codex)`  [EXTRACTED] [semantically similar]
  .claude/skills/repomind-query/SKILL.md → .codex/skills/repomind-query/SKILL.md
- `repomind-summary Skill` --semantically_similar_to--> `repomind-summary Skill (Codex)`  [EXTRACTED] [semantically similar]
  .claude/skills/repomind-summary/SKILL.md → .codex/skills/repomind-summary/SKILL.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Container Lifecycle Orchestration in main** — main_go_main, network_network_init, cgroups_cgroup_configdefaultcgroups, network_network_configdefaultnetworkinnewnet, cgroups_cgroup_cleancgroupspath, workspace_workspace_delmntnamespace, workspace_workspace_setmntnamespace, network_network_waitparentsetnewnet [INFERRED 0.95]
- **RepoMind Query-Summary Mandatory Pair** — claude_rules_repomind_repomind_query, claude_rules_repomind_repomind_summary, claude_rules_repomind_summary_gate, claude_rules_repomind_frontmatter_routing [EXTRACTED 1.00]
- **Filesystem-Backed Network State Persistence** — network_network_netmgr, network_ipam_fs_ipamfs, config_config_netstoragepath, config_config_ipamstoragefspath, network_bitmap_bitmap [INFERRED 0.95]
- **RepoMind Skill Pipeline: init -> prd -> query -> summary** — repomind_init_skill_repomind_init, repomind_prd_skill_repomind_prd, repomind_query_skill_repomind_query, repomind_summary_skill_repomind_summary [EXTRACTED 1.00]
- **RepoMind Knowledge Kinds** — repomind_knowledge_types, repomind_frontmatter_routing, repomind_query_findings [EXTRACTED 1.00]
- **Tinydocker Containerization Concepts** — readme_tinydocker_project, readme_containerization_principles [EXTRACTED 1.00]

## Communities (12 total, 3 thin omitted)

### Community 0 - "Container Lifecycle"
Cohesion: 0.17
Nodes (16): CleanCgroupsPath(), ConfigDefaultCgroups(), Banner(), Debug (package-level), defaultLogger, Info (package-level), Warn (package-level), main (+8 more)

### Community 1 - "RepoMind Skills"
Cohesion: 0.17
Nodes (20): RepoMind Rules (AGENTS.md), Code Search Priority Chain, Frontmatter-based Routing, graph-scan Command, repomind-init Skill (Codex), kb-metadata Command, kb-migrate Command, Knowledge Types: concepts, modules, troubles (+12 more)

### Community 2 - "Bridge Network Driver"
Cohesion: 0.19
Nodes (15): Link, Error (package-level), NetConf, createBridge(), enterContainerNetns(), genInterfaceIp(), IP, IPNet (+7 more)

### Community 3 - "Logging System"
Cohesion: 0.18
Nodes (6): ColorLogger, InitWriteLogger(), New(), Logger, WaitGroup, Writer

### Community 4 - "IPAM Filesystem"
Cohesion: 0.32
Nodes (9): bitMap, IpAmStorageFsPath, IPMask, InitBitMap(), getIPIndex(), IP, ipToUint32(), uint32ToIP() (+1 more)

### Community 5 - "Network Tests"
Cohesion: 0.26
Nodes (8): bitMap, arrIndex(), bytePos(), T, TestBitSet(), T, TestAlloc(), TestBitMap_BitClean()

### Community 6 - "Workspace Namespaces"
Cohesion: 0.57
Nodes (7): delMntNamespace(), delMntNamespace, mntLayer(), mntOldLayer(), SetMntNamespace(), workerLayer(), writeLayer()

### Community 7 - "Network Manager Config"
Cohesion: 0.33
Nodes (4): NetStoragePath, NetConf, netMgr, IPNet

### Community 8 - "RepoMind Rules"
Cohesion: 0.40
Nodes (5): Frontmatter Metadata Routing, repomind-query, repomind-summary, Summary Gate, Graphify Full Rebuild

## Knowledge Gaps
- **24 isolated node(s):** `WaitGroup`, `T`, `networktype`, `Veth`, `NetConf` (+19 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `ipAmFs` connect `IPAM Filesystem` to `Container Lifecycle`, `Network Tests`?**
  _High betweenness centrality (0.234) - this node is a cross-community bridge._
- **Why does `defaultLogger` connect `Container Lifecycle` to `Bridge Network Driver`, `Logging System`?**
  _High betweenness centrality (0.222) - this node is a cross-community bridge._
- **Why does `InitWriteLogger()` connect `Logging System` to `Container Lifecycle`?**
  _High betweenness centrality (0.170) - this node is a cross-community bridge._
- **What connects `WaitGroup`, `T`, `networktype` to the rest of the system?**
  _33 weakly-connected nodes found - possible documentation gaps or missing edges._