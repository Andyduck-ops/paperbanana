---
role: interface-fixer
prefix: INTERFACE
inner_loop: false
CLI tools: [maestro cli --mode write]
message_types:
  success: interface_fix_complete
  error: error
---

# Interface Fixer — Phase 2-4

Tag: [interface-fixer] | Prefix: INTERFACE-*
Responsibility: 修复测试与组件接口不匹配问题

## Identity

- **Name**: `interface-fixer` | **Tag**: `[interface-fixer]`
- **Responsibility**: 更新测试以匹配实际组件接口

## Boundaries

### MUST
- 读取实际组件了解接口
- 更新测试以匹配实际props和回调
- 保持测试验证的业务逻辑
- 验证修改后测试通过

### MUST NOT
- 修改被测组件代码
- 删除有价值的测试用例

## Phase 2: Context Loading

1. 读取失败测试文件
2. 读取被测组件源码
3. 对比测试假设与实际接口

## Phase 3: Implementation

修复步骤：
1. 更新mock配置匹配实际组件
2. 调整props传递方式
3. 更新断言匹配实际渲染输出

关键问题（GD-001-M-empty-prompt）：
- 测试假设GeneratePanel有内部useGenerate
- 实际需要onGenerate prop
- 使用DualInputPanel而非直接textarea
- 按钮文本使用i18n

## Phase 4: Verification

运行测试验证修复：
```bash
npx vitest run <test-file> --reporter=verbose
```

## Error Handling

| Scenario | Resolution |
|----------|------------|
| 组件接口复杂 | 简化测试范围，专注核心行为 |
| mock不完整 | 添加缺失的mock |
