# Introduction to mdex Documentation

This section provides an overview of how to structure your documentation for
`mdex`.

## Document Structure

`mdex` processes Markdown files and organizes them based on your directory
structure. Each Markdown file (`.md`) is converted into an HTML page.

### Headings and TOC

Headings (H1, H2, H3, etc.) are automatically used to generate a Table of
Contents (TOC) on each page. For example:

```markdown
# Main Topic

## Sub-Topic 1

### Detail 1.1

## Sub-Topic 2
```

This structure will create a navigable TOC on the right side of your generated
page.

## Code Blocks

You can include code blocks with syntax highlighting:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, mdex!")
}
```

## Links

Internal links to other Markdown files should use relative paths and refer to
the `.html` extension that `mdex` will generate:

- [Back to Home](../index.html)
- [Daily Notes](../../notes/daily.html)

External links work as usual:

- [Goldmark GitHub](https://github.com/yuin/goldmark)

### Autolinks

URLs and email addresses are automatically converted into links:

- Visit our website: https://example.com
- Contact us at: info@example.com

## GFM Features

### Strikethrough

You can easily ~~strike through~~ text.

### Task Lists

- [x] Completed task
- [ ] Pending task

## Images

Images can be included using standard Markdown syntax. Ensure your image paths
are correct relative to the Markdown file.

![Example Image](https://via.placeholder.com/150)

## Lists

### Unordered List

- Item 1
- Item 2
  - Sub-item 2.1
  - Sub-item 2.2

### Ordered List

1.  First item
2.  Second item
    1.  Nested ordered item
    2.  Another nested item

## Tables

| Header 1    | Header 2    |
| ----------- | ----------- |
| Row 1 Col 1 | Row 1 Col 2 |
| Row 2 Col 1 | Row 2 Col 2 |

This concludes the introduction. Explore more examples in the `notes` directory!