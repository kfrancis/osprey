Here's a complete design system for a modern programming language website based on the colors from your logo. The colors are extracted from the osprey illustration and structured to guide you across UI elements.

---

## 🎨 **Color Palette**

**Primary Colors**

* `BlueDark`: `#0B3A5E` — used for text, headers, nav background
* `BlueMid`: `#005C9C` — for buttons, accents, links
* `BlueLight`: `#38A1DB` — call to action, highlights

**Secondary Colors**

* `Aqua`: `#77D7F4` — cards, hover states, subtle fills
* `Sky`: `#C7EBF9` — background sections, borders
* `OffWhite`: `#FDFDFD` — page background

**Neutral**

* `DarkGrey`: `#1F1F1F` — body text
* `LightGrey`: `#F2F2F2` — containers, lines, code blocks
* `White`: `#FFFFFF` — cards, panels

---

## 🔤 **Typography**

* **Font Family**: `Inter`, `SF Pro Text`, `system-ui`, sans-serif
* **Weights**: 400 (body), 500 (emphasis), 700 (headings)

### Sizes

* `Display`: 64px
* `Headline`: 40px
* `Subhead`: 24px
* `Body`: 16px
* `Code`: 14px (monospace: `JetBrains Mono`, `Fira Code`)

---

## 🧱 **Layout & Spacing**

* **Max Width**: 1200px
* **Gutters**: 24px
* **Grid**: 12-column flexible grid

**Spacing Scale (px)**:
`4, 8, 16, 24, 32, 48, 64, 96`

---

## 🧩 **Components**

### Button

```css
.button-primary {
  background: #005C9C;
  color: #FFFFFF;
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  font-weight: 500;
  transition: background 0.2s ease;
}
.button-primary:hover {
  background: #007ACC;
}
```

### Code Block

```css
.code-block {
  background: #F2F2F2;
  border-left: 4px solid #005C9C;
  font-family: 'JetBrains Mono', monospace;
  font-size: 14px;
  padding: 16px;
  border-radius: 6px;
}
```

### Card

```css
.card {
  background: #FFFFFF;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(11, 58, 94, 0.08);
  padding: 24px;
  transition: box-shadow 0.2s ease;
}
.card:hover {
  box-shadow: 0 8px 24px rgba(11, 58, 94, 0.15);
}
```

---

## 🔄 **Interactions**

* **Transitions**: `all 0.2s ease` for buttons, cards, links
* **Hover States**: slightly lighter or darker tone
* **Focus States**: `outline: 2px solid #77D7F4`

---

## 🧠 **Usage Guidelines**

* Keep the blue gradients consistent — darker for navs and actions, lighter for backgrounds.
* Use `Aqua` or `Sky` sparingly to guide the eye or divide sections.
* All icons should follow a single-weight stroke (24px line icon set).
* Use whitespace generously. Let the brand breathe.
