# ADR-001: UI Design Philosophy

## Status

Accepted

## Context

The Skeleton CRM user interface needs to reflect the system's philosophy: enterprise-grade, professional, efficient, and focused on productivity. Unlike many modern SaaS applications with colorful gradients, illustrations, and playful elements, Skeleton CRM should embody minimalist elegance.

## Decision

We adopt a **"Monochrome Cinema"** design philosophy inspired by classic black-and-white films:

### Core Principles

1. **High Information Density**
   - Dense, information-rich layouts
   - No wasted space
   - Every pixel serves a purpose
   - Tables over cards when data is primary

2. **Monochromatic Color Palette
   - Base: Grayscale (#000 to #FFF)
   - Primary: Neutral grays for hierarchy
   - Accent: Muted blue (#3B82F6) only for active/focus states
   - Success/Error: Muted green/red for semantic meaning
   - NO bright colors, NO gradients, NO illustrations

3. **Typography-Driven UI
   - Text is the primary interface element
   - Clear typographic hierarchy
   - Monospace for data/numbers
   - Sans-serif for UI elements
   - Limited font sizes (SM, Base, LG, XL, 2XL)

4. **Subtle Interactions
   - NO animations, transitions, or micro-interactions
   - Instant state changes
   - Static, predictable layouts
   - Keyboard-first navigation

5. **Functional Minimalism
   - NO decorative images or icons
   - NO hero illustrations
   - NO empty states with illustrations
   - NO loading spinners (use skeletal loading)
   - Icons only when semantically necessary (semantic icons)

### Component Patterns

#### Forms
- Compact, vertical layouts
- Labels above inputs
- Minimal padding
- Error messages inline
- NO floating labels

#### Tables
- Dense rows (h-10)
- Minimal borders
- Hover state only
- Zebra striping optional
- Sortable columns with subtle indicators

#### Navigation
- Sidebar-based (not top nav)
- Collapsible sections
- Flat hierarchy
- Icon + text labels
- Active state: background change, no borders

#### Authentication
- Centered, compact forms
- NO welcome illustrations
- NO social login buttons (unless required)
- Clean error states
- Focus on efficiency

### Design References

- **shadcn/ui Blocks** - Base component patterns
- **Linear** - Dense, calm interface
- **Paperless-NG** - Document-focused UI
- **Gmail/Google Docs** - Information density
- **Terminal/IDE** - Focus on content

## Implementation

### Tailwind Config

```typescript
// tailwind.config.ts
const config = {
  theme: {
    extend: {
      colors: {
        // Monochromatic scale
        monochrome: {
          50: '#FAFAFA',
          100: '#F4F4F5',
          200: '#E4E4E7',
          300: '#D4D4D8',
          400: '#A1A1AA',
          500: '#71717A',
          600: '#52525B',
          700: '#3F3F46',
          800: '#27272A',
          900: '#18181B',
          950: '#09090B',
        },
      },
    },
  },
}
```

### CSS Variables

```css
:root {
  --background: 0 0% 100%;
  --foreground: 240 10% 3.9%;
  --card: 0 0% 100%;
  --card-foreground: 240 10% 3.9%;
  --primary: 240 5.9% 10%;
  --primary-foreground: 0 0% 98%;
  --secondary: 240 4.8% 95.9%;
  --muted: 240 4.8% 95.9%;
  --muted-foreground: 240 3.8% 46.1%;
  --accent: 240 4.8% 95.9%;
  --border: 240 5.9% 90%;
}
```

### Component Examples

#### Login Form (Monochrome Cinema)

```tsx
// Dense, compact, typography-focused
<div className="flex flex-col space-y-4">
  <div className="space-y-1">
    <h1 className="text-2xl font-semibold tracking-tight">
      Sign in
    </h1>
    <p className="text-sm text-muted-foreground">
      Enter your credentials to continue
    </p>
  </div>
  
  <form className="space-y-3">
    <div className="space-y-1">
      <Label htmlFor="email" className="text-xs font-medium">
        Email
      </Label>
      <Input id="email" type="email" className="h-9" />
    </div>
    
    <div className="space-y-1">
      <Label htmlFor="password" className="text-xs font-medium">
        Password
      </Label>
      <Input id="password" type="password" className="h-9" />
    </div>
    
    <Button type="submit" className="h-9 w-full">
      Sign in
    </Button>
  </form>
  
  <p className="text-xs text-muted-foreground">
    Don't have an account?{' '}
    <Link href="/register" className="underline">
      Create one
    </Link>
  </p>
</div>
```

## Consequences

### Positive

- Fast rendering (minimal CSS/JS)
- Clear visual hierarchy
- Accessible by default
- Professional appearance
- Focus on data and tasks
- Easy to maintain
- Works on all devices/bandwidths

### Negative

- May appear "boring" to some users
- Requires strong typography skills
- Less emotional engagement
- Differentiates through function, not form

## Alternatives Considered

1. **Colorful + Illustrations** - Rejected: Distracts from data, slower load
2. **Gradient-based** - Rejected: Not enterprise-appropriate
3. **Minimalist White-space** - Rejected: Wastes screen real estate

## References

- [shadcn/ui Blocks](https://ui.shadcn.com/blocks)
- [Linear Design Principles](https://linear.app)
- [Refactoring UI](https://refactoringui.com/)

---

**Author:** basilex  
**Date:** 2026-04-09  
**Supersedes:** N/A