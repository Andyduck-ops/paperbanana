# Coordinate Step 1/3 — ui-refactor-driven

Execute the following command:

/spec-add "添加 UI 架构重构 Golden Data 规范：模型配置、精修迭代、批量变体"

Auto-confirm all prompts. No interactive questions.

## 具体要求

基于我们刚刚讨论的 UI 架构设计 V2.1，需要添加以下 Golden Data 规范：

### 1. GD-UI-006: 模型配置可见性
- 用户可以查看当前使用的检索模型和生图模型
- 用户可以在 Popover 中切换渠道和模型
- 商家可以启用/禁用
- 模型可以通过 +/- 添加/移除

### 2. GD-UI-007: 精修迭代进度
- 支持多轮迭代精修 (1-5 轮)
- 每轮显示 Before/After 对比
- 显示 AI 描述的改动说明
- 支持从任意轮继续迭代

### 3. GD-UI-008: 批量变体展示
- 支持批量生成多种风格变体 (1-10 种)
- 变体以网格形式展示
- 可以选中变体继续迭代
- 支持下载全部或选择下载

### 4. GD-UI-009: 极简状态展示
- 底部输入框上方显示极简进度条
- 进度条样式类似 Vercel 部署进度
- 失败时单行红色文本 + 简短原因
- 不使用大卡片堆砌进度

### 5. GD-UI-010: 悬浮输入区
- 输入框悬浮在屏幕底部中央
- 带精致阴影和毛玻璃效果
- 支持 Method/Caption 双输入
- 生成时显示 Stop 按钮

请将这些规范写入 `test/golden/cases/ui/` 目录下的 YAML 文件中。

## Return Format

Output MUST end with this exact block (fill each field on its own line):

```
--- COORDINATE RESULT ---
STATUS: <SUCCESS or FAILURE>
PHASE: <number, or "none">
ARTIFACTS: <comma-separated file paths, or "none">
SUMMARY: <one-line what was accomplished, include session IDs if any>
```

Rules:
- Execute the command as-is — it is a maestro slash command with arguments
- Do not modify files outside the command's intended scope
- PHASE must reflect the phase number referenced during execution
