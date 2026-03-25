# UI Text Audit Report - PaperBanana Clean Frontend

**Task ID**: RESEARCH-001
**Date**: 2026-03-25
**Project**: paperbanana-clean/web
**Reference**: repo-cn (Python CLI project with static HTML landing page)

## Executive Summary

The paperbanana-clean frontend has a well-structured i18n (internationalization) system using react-i18next with Chinese (zh) and English (en) locale files. However, several hardcoded strings remain in components, particularly in error messages, empty states, and accessibility labels. The UI text is generally professional and suitable for an academic paper figure generation tool.

---

## 1. i18n Infrastructure Analysis

### 1.1 Current State

**Location**: `web/src/i18n/`

| File | Purpose |
|------|---------|
| `index.ts` | i18next configuration, language switching |
| `types.d.ts` | TypeScript type definitions for translations |
| `locales/zh.json` | Chinese translations (398 lines) |
| `locales/en.json` | English translations (398 lines) |

**Configuration**:
- Default language: Chinese (`zh`)
- Fallback language: English (`en`)
- Two supported languages with native names

### 1.2 Translation Coverage

**Well-covered areas**:
- App name and branding
- Theme options (4 themes: Qi Baishi, Pop Anime, Rococo, Night Mono)
- Generation panel (stages, modes, pipelines)
- Settings page (providers, channels, models)
- History panel
- Export modal
- Validation messages
- Common UI elements (buttons, actions)

**Missing/incomplete areas**:
- Error boundary messages
- Empty state examples
- Some accessibility labels

---

## 2. Hardcoded Text Findings

### 2.1 Critical Issues (Should be internationalized)

#### ErrorBoundary.tsx (Line 27-34)
```tsx
<h2 className="text-xl text-foreground mb-4">Something went wrong</h2>
<p className="text-muted-foreground mb-4">{this.state.error?.message}</p>
<button ...>Reload Page</button>
```
**Recommendation**: Add to `error` namespace in locale files.

#### EmptyState.tsx (Lines 116-123, 130, 179-181)
```tsx
// Line 116-117
'Create Scientific Visualizations'
'Refine Your Images'

// Line 120-122
'Transform your research into publication-ready figures...'
'Upload an image and provide refinement instructions...'

// Line 130
'Try an example to get started'

// Line 179-181
'Or type your own description below to begin'
'Upload an image below to start refining'
```
**Note**: The i18n keys exist in `workspace.emptyState` namespace but are not being used. This appears to be incomplete implementation.

#### DualInputPanel.tsx (Lines 76, 123)
```tsx
// Line 76, 123
{methodContent || <span className="text-muted-foreground">No content</span>}
{caption || <span className="text-muted-foreground">No content</span>}
```
**Recommendation**: Add `common.noContent` key.

#### ImageUpload.tsx (Line 87)
```tsx
alt="Uploaded preview"
```
**Recommendation**: Add to `refine.uploadedPreview` key.

#### CandidateGrid.tsx (Lines 114, 126)
```tsx
aria-label="Grid view"
aria-label="List view"
```
**Recommendation**: Add `batch.gridView` and `batch.listView` keys.

#### ModeSwitcher.tsx (Line 34)
```tsx
aria-label="Workspace mode"
```
**Recommendation**: Add `workspace.modeLabel` key.

### 2.2 Example Content (Intentionally English)

#### GeneratePanel.tsx (Lines 23-31)
```tsx
const DEFAULT_EXAMPLES = [
  {
    method: `Database retrieval notes:...`,
    caption: 'Generate a clean process diagram...',
  },
];
```
**Assessment**: This is sample content for demonstration purposes. Should provide localized examples for each language.

#### EmptyState.tsx (Lines 11-35)
```tsx
const EXAMPLE_PROMPTS = [
  {
    id: 'architecture',
    title: 'Neural Network Architecture',
    description: 'Visualize transformer attention patterns',
    prompt: 'Create a detailed diagram...',
  },
  // ... more examples
];
```
**Assessment**: Should be internationalized or moved to locale files.

---

## 3. UI Text Quality Assessment

### 3.1 Correctness (No typos/semantic issues found)

All UI text in the locale files appears correct:
- No spelling errors detected
- Semantic clarity is good
- Professional tone maintained

### 3.2 Example Code/Data Assessment

| Component | Status | Notes |
|-----------|--------|-------|
| DEFAULT_EXAMPLES in GeneratePanel | **Needs localization** | Only English examples provided |
| EXAMPLE_PROMPTS in EmptyState | **Needs localization** | Only English examples provided |
| Workspace examples in zh.json | Present | 4 example categories defined |

### 3.3 Error Message Quality

**Current state**:
| Error Type | Quality | Recommendation |
|------------|---------|----------------|
| Network errors | Good | Clear, actionable |
| Server errors | Good | Simple, helpful |
| Validation errors | Good | Specific to field |
| ErrorBoundary | **Needs improvement** | Generic, hardcoded |

**Sample from locale files**:
```json
"error": {
  "networkError": "Network error. Please check your connection.",
  "serverError": "Server error. Please try again later.",
  "validationError": "Please check your input."
}
```

---

## 4. Comparison with repo-cn Reference

### 4.1 Key Differences

| Aspect | paperbanana-clean | repo-cn |
|--------|-------------------|---------|
| Architecture | React SPA | Python CLI + Static HTML |
| i18n | Full i18next system | No i18n (English only) |
| UI Focus | Interactive workspace | Landing page for paper |
| Text scope | Full application | Marketing/research content |

### 4.2 Terminology Alignment

The paperbanana-clean UI terminology aligns with repo-cn:

| Term | paperbanana-clean | repo-cn |
|------|-------------------|---------|
| PaperBanana | Correct | Correct |
| Retriever Agent | Referenced | Defined |
| Planner Agent | Referenced | Defined |
| Visualizer Agent | Referenced | Defined |
| Stylist Agent | Referenced | Defined |
| Critic Agent | Referenced | Defined |

---

## 5. Accessibility Text Review

### 5.1 aria-label Audit

| Component | Current State | Recommendation |
|-----------|---------------|----------------|
| ModeSwitcher | Hardcoded "Workspace mode" | Internationalize |
| CandidateGrid view toggles | Hardcoded "Grid view"/"List view" | Internationalize |
| Header theme buttons | Dynamic, good | OK |
| Header language buttons | Dynamic, good | OK |

### 5.2 Screen Reader Announcements

Good use of aria-live regions in:
- Progress announcements (`generate.announcementRunning`, etc.)
- Status updates
- Toast notifications

---

## 6. Recommendations Summary

### High Priority

1. **Fix EmptyState component** - Use existing i18n keys instead of hardcoded strings
2. **Internationalize ErrorBoundary** - Add error messages to locale files
3. **Internationalize accessibility labels** - All aria-labels should use t()

### Medium Priority

4. **Localize example content** - Provide language-specific examples
5. **Add "No content" translation** - For empty preview states
6. **Localize alt text** - Image alt attributes should be translated

### Low Priority

7. **Consider adding more detailed error messages** - For common failure scenarios
8. **Add tooltips for complex settings** - Help text for configuration options

---

## 7. Implementation Checklist

### New i18n Keys Required

```json
{
  "common": {
    "noContent": "No content",
    "reload": "Reload Page"
  },
  "error": {
    "genericTitle": "Something went wrong",
    "genericMessage": "An unexpected error occurred. Please try again."
  },
  "workspace": {
    "modeLabel": "Workspace mode"
  },
  "batch": {
    "gridView": "Grid view",
    "listView": "List view"
  },
  "refine": {
    "uploadedPreview": "Uploaded preview"
  },
  "emptyState": {
    "tryExample": "Try an example to get started",
    "orTypeBelow": "Or type your own description below to begin",
    "uploadBelow": "Upload an image below to start refining"
  }
}
```

---

## 8. Files Requiring Modification

| File | Changes Needed |
|------|----------------|
| `components/ErrorBoundary.tsx` | Replace hardcoded text with t() calls |
| `components/workspace/EmptyState.tsx` | Use existing i18n keys, add missing keys |
| `components/DualInputPanel.tsx` | Add "No content" translation |
| `components/ImageUpload.tsx` | Translate alt text |
| `components/workspace/CandidateGrid.tsx` | Internationalize aria-labels |
| `components/workspace/ModeSwitcher.tsx` | Internationalize aria-label |
| `components/GeneratePanel.tsx` | Move examples to locale or create language-specific examples |
| `i18n/locales/zh.json` | Add new keys |
| `i18n/locales/en.json` | Add new keys |

---

## Appendix: Tech Stack

- **Framework**: React 19.2.4
- **Build**: Vite 6.0.0
- **i18n**: react-i18next 16.5.8 / i18next 25.8.18
- **Styling**: Tailwind CSS 4.2.1
- **Testing**: Vitest 3.2.4, Testing Library

---

**End of Report**
