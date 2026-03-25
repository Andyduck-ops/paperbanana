# PaperBanana Frontend Workspace Rebuild PRD

## 1. Document Positioning

This PRD defines the next-stage frontend rebuild direction for PaperBanana.
It is not a visual mockup file and not an implementation checklist.
Its purpose is to align product intent before large-scale UI execution, multi-agent implementation, and follow-up Golden Data expansion.

The current rebuild target is not "polish the page a little".
It is to reconstruct PaperBanana into a coherent academic paper figure workspace with clearer information architecture, stronger model configuration capability, and a more intentional visual language.

## 2. Background and Problem Statement

The current frontend has four structural problems:

1. Core workspace focus is weak.
   History, settings, generation, and refinement are not yet arranged into a unified creation workspace.
2. Model configuration capability is fragmented.
   The UI does not yet provide a complete and unified way to manage channels, fetch models inside channels, and map models to different workflow roles.
3. Product extensibility is constrained.
   Refinement flow and batch generation are not yet expressed as first-class parts of the same workspace grammar.
4. Visual language is underdefined.
   The interface can still be substantially upgraded in rhythm, atmosphere, and design identity.

This PRD aims to solve these by defining a single coherent frontend workspace model.

## 3. Product Goal

PaperBanana should evolve into an academic paper figure generation and refinement workspace with the following properties:

- the main workspace remains visually dominant
- history is available but can be neatly stored away
- model configuration is unified under a channel-model system
- advanced model selection is possible without locking the whole workflow to one provider
- generation and refinement are two modes of one workspace, not two disconnected products
- multiple candidates can be generated and compared in one view
- theme switching becomes part of the product's aesthetic identity, not a technical dropdown

## 4. Non-Goals

The current phase does not aim to solve these in full:

- per-agent free model routing for every pipeline stage
- deep provider capability benchmarking UI
- full enterprise settings center
- pixel-perfect final visual polish for every component
- complete design system standardization across the whole repository in one pass

These may happen later, but they are not the scope of this PRD.

## 5. Target Users

Primary users:

- users generating academic paper figures, scientific diagrams, or plots from natural language prompts
- users iterating on generated outputs through refinement
- advanced users who need to manage multiple model channels and choose different models for different workflow roles

User traits:

- they care about output quality and iteration efficiency
- they need clarity, not decorative complexity
- they may switch frequently between generation, review, and refinement
- they benefit from history as a semantic recall tool, not as a file explorer

## 6. Core Product Principles

### 6.1 Main Stage First

The central workspace is the primary stage.
Everything else should serve it.
History, settings, and theme switching should never visually overpower the current creative task.

### 6.2 Configuration Must Be Structured

Model configuration should move from scattered form fields to a clear system:
channel -> models inside channel -> workflow role mapping.

### 6.3 One Workspace, Multiple Modes

Generation and refinement should be expressed as two modes inside the same workspace grammar.
The product should feel continuous rather than fragmented.

### 6.4 History Is Semantic Memory

History should not be a raw session dump.
It should function as retrievable memory with timestamp plus semantic summary.

### 6.5 Themes Are Visual Identity, Not Technical Toggles

Theme selection should be designed as a visual choice through colored marks, not textual labels like "Theme A" or "Light".

## 7. Information Architecture

The rebuilt frontend workspace is composed of five major areas:

1. Main Workspace
2. History Panel
3. Unified Model Configuration Area
4. Mode Switch Area
5. Theme Selector

### 7.1 Main Workspace

The main workspace is the dominant visual zone.
It hosts:

- generation input
- refinement input
- in-progress state feedback
- candidate result comparison
- final result actions

This area should remain the focal point at all times.

### 7.2 History Panel

History is stored in a left-side sliding panel.
It is not permanently expanded.
It must have a clear storage point and a clear recall point.

The panel should display:

- timestamp
- semantic summary title
- run status
- mode hint when useful

The semantic summary title should take inspiration from Cherry Studio style topic naming:
short, intention-centered, and recall-friendly.

### 7.3 Unified Model Configuration Area

The model configuration experience should be unified under a single design language.
It should support:

- multiple channels
- channel-level configuration
- fetching models inside a channel
- adding or removing models within a channel via plus/minus style controls
- mapping one model to image generation
- mapping one model to retrieval/reasoning

This is the core structural upgrade of the settings experience.

### 7.4 Mode Switch Area

Generation and refinement should be two explicit workspace modes.
They should be switched lightly and directly, without making the user feel they have navigated into a different product.

### 7.5 Theme Selector

Theme switching should use four color-coded marks instead of alphabetic naming.
The selector should feel like a visual palette, not a settings dropdown.

## 8. Functional Requirements

## 8.1 History Panel Requirements

### FR-H1

The interface must provide a left-side storage point for history access.

### FR-H2

Clicking the history entry point must open a left sliding panel.

### FR-H3

The panel must be closable and support returning focus to the main workspace.

### FR-H4

Each history item must display:

- time
- semantic summary title
- status

### FR-H5

History naming should prioritize task intent summary rather than generic session numbering.

### FR-H6

Selecting a history item should restore that session into the workspace flow.

## 8.2 Unified Model Configuration Requirements

### FR-M1

The product must provide one unified model configuration surface.

### FR-M2

Users must be able to add multiple channels.

### FR-M3

Each channel must support entering channel-level configuration data.

### FR-M4

Each channel must support fetching or listing available models inside that channel.

### FR-M5

Users must be able to add or remove candidate models within each channel through a clear plus/minus interaction pattern.

### FR-M6

The product must support assigning:

- one image generation model
- one retrieval/reasoning model

where each assignment is expressed as `channel + model`.

### FR-M7

The model configuration UI must not assume that the whole product is locked to one provider.

### FR-M8

Model configuration acts as a global workspace state. However, each generated artifact must retain a snapshot of the `channel + model` pair used, ensuring historical reproducibility even if global settings change.

## 8.3 Advanced Configuration Requirements

### FR-A1

Advanced configuration must support selecting image generation and retrieval/reasoning from different channels.

### FR-A2

The advanced configuration experience should remain structurally simple in this phase and must not expose full per-agent routing.

### FR-A3

The user should be able to clearly understand which `channel + model` pair is currently mapped to each workflow role.

## 8.4 Generation and Refinement Workflow Requirements

### FR-W1

Generation and refinement must exist as two first-class workspace modes.

### FR-W2

Refinement should be able to participate in iterative pipeline flow in future extensions.

### FR-W3

Refinement should support multi-batch candidate generation capability through reuse of the existing generation/batch logic where possible.

### FR-W4

When multiple candidates are generated, the user must be able to view them together in one result area.

### FR-W5

The result area should support selecting one candidate as the preferred branch for follow-up refinement or export.

## 8.5 Theme Requirements

### FR-T1

The product must provide four themes.

### FR-T2

Theme selection must use non-textual color marks as the primary selector.

### FR-T3

Each theme must have a distinct visual direction and should be described by design style rather than letter naming.

## 9. Interaction Design Requirements

## 9.1 History Interaction

- history enters from the left as a sliding panel
- the panel should feel secondary to the main stage
- opening and closing should be smooth and reversible
- history should act like stored context, not like a permanent dashboard

## 9.2 Model Configuration Interaction

- configuration should be unified rather than scattered across unrelated areas
- channel management and model management should feel structurally related
- adding/removing models inside a channel should be direct and low-friction
- workflow role mapping should be easy to read at a glance

## 9.3 Mode Switching

- switching between generation and refinement should feel like mode change, not page jump
- the user should preserve mental continuity across modes

## 9.4 Candidate Comparison

- multiple candidates should be displayed together
- the comparison layout should allow quick scanning
- users should be able to promote one candidate into the main working branch

## 10. Visual and Aesthetic Direction

The current frontend may be substantially restructured.
The new visual direction should avoid generic dashboard aesthetics.
It should be treated as a focused creation workspace with a stronger art and design identity.

The four theme directions for this phase are:

1. Swiss International Style

   - strong grid logic
   - rational composition
   - restrained contrast
   - information order first

2. Bauhaus Functionalism

   - geometric modularity
   - function-led composition
   - clear structure
   - disciplined color use

3. Neo-Minimal Naturalism

   - softer surfaces
   - calm low-saturation palette
   - long-session friendliness
   - academic warmth

4. Art Deco Digital Elegance
   - refined ornamental restraint
   - stronger formal identity
   - elevated finish
   - selective theatricality

These labels are directional anchors for visual language.
They are not implementation-level color tokens.

## 11. Empty State Requirements

The rebuilt workspace should define an intentional empty state.
When the user opens the product without an active session, the interface should not feel hollow.

The empty state should include:

- a brief statement of capability
- sample task prompts or suggested entry points
- a clear invitation into generation mode

This is necessary to reduce first-use friction and establish product tone.

## 12. State Feedback Requirements

The product should maintain a consistent tone when expressing progress, failure, and completion.
State feedback should be:

- concise
- professional
- understandable
- non-gimmicky

It should avoid both robotic emptiness and over-personified chatter.

## 13. Result Area Requirements

The result area should distinguish between:

- current preferred result
- alternative candidates
- next-step actions such as refine or export

The result area should not flatten all outputs into one undifferentiated card stack.

## 14. Golden Data Alignment

This PRD must remain compatible with the current UI Golden Data contract, especially:

- `test/golden/cases/ui/GD-UI-001.yaml`
- `test/golden/cases/ui/GD-UI-002.yaml`
- `test/golden/cases/ui/GD-UI-003.yaml`
- `test/golden/cases/ui/GD-UI-004.yaml`
- `test/golden/cases/ui/GD-UI-005.yaml`

This means the rebuilt UI must still preserve:

- visible multi-stage progress
- visible failure location
- surfaced final artifact
- resumed-task semantics
- per-task batch visibility where applicable

## 15. Success Criteria

This rebuild is successful if:

1. users can access history without history permanently dominating screen space
2. model configuration becomes unified, understandable, and extensible
3. image generation and retrieval/reasoning can use different `channel + model` assignments
4. refinement is structurally prepared to share generation and batch logic
5. theme switching becomes a deliberate visual interaction rather than a generic settings field
6. the workspace feels like one coherent creative product rather than several disconnected pages

## 16. Future Extensions

Future iterations may add:

- deeper per-stage model routing
- manual history naming and grouping
- richer candidate comparison modes
- advanced refinement loop controls
- broader design system tokenization

These are explicitly outside the current PRD scope.


测试key：sk-O3vtBPphF7UPLkmGlNBc0qt5jWrkPy4H2ApWHUDeQDU0YhaQ //测试的URL：https://lx.lxsummer.cloud/ 模型生图什么的都使用grok相关的
