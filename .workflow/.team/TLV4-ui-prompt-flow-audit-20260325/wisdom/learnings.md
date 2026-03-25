# Learnings - UI Prompt Flow Audit Session

## 2026-03-25: Prompt Template Audit

### Key Findings

1. **Style Guide Compression Issue**
   - Go's Stylist agent uses a ~50 word inline style guide summary
   - Python loads comprehensive ~2000 word NeurIPS 2025 guides from files
   - This difference could significantly affect diagram/plot aesthetic quality
   - Recommendation: Expand Go's style guide to match Python's comprehensiveness

2. **Token Efficiency vs Quality Trade-off**
   - Go enforces strict token limits (3K chars for target, 1.6K for examples)
   - Python has no limits, potentially exceeding context windows
   - Go's pre-scoring in Retriever is a good optimization
   - Balance needed between efficiency and quality

3. **Version Control Improvement**
   - Go has `PromptVersion` constants for all agents
   - Python lacks versioning, making prompt evolution tracking harder
   - This is a good practice that should be backported to Python

4. **Architecture Pattern**
   - Go's callback-based design (loadImage, PlotExecutor) is cleaner
   - Allows better testing and mocking
   - Python's inline implementation is more coupled

### Action Items for Future Work

- [ ] Expand style guide in Go Stylist agent
- [ ] Add lite/full mode toggle to Go Retriever
- [ ] Standardize output key naming between implementations
- [ ] Make token limits configurable in Go
