---
title: List render hooks
linkTitle: Lists
description: Create list render hook templates to override the rendering of Markdown lists and list items to HTML.
categories: []
keywords: []
---

{{< new-in 0.XXX.0 />}}

## render-list

Use a `render-list` hook template to override the rendering of Markdown lists.

### Context

List _render hook_ templates receive the following [context](g):

Attributes
: (`map`) The [Markdown attributes], available if you configure your site as follows:

  {{< code-toggle file=hugo >}}
  [markup.goldmark.parser.attribute]
  block = true
  {{< /code-toggle >}}

IsOrdered
: (`bool`) Returns `true` if the list is an ordered list (rendered as `<ol>`), or `false` if it is an unordered list (rendered as `<ul>`).

Ordinal
: (`int`) The zero-based ordinal of the list on the page.

Page
: (`page`) A reference to the current page.

PageInner
: (`page`) A reference to a page nested via the [`RenderShortcodes`] method. [See details](#pageinner-details).

Position
: (`string`) The position of the list within the page content.

Text
: (`template.HTML`) The rendered list items.

[Markdown attributes]: /content-management/markdown-attributes/
[`RenderShortcodes`]: /methods/page/rendershortcodes

## render-listitem

Use a `render-listitem` hook template to override the rendering of individual Markdown list items.

### Context

List item _render hook_ templates receive the following [context](g):

Ordinal
: (`int`) The zero-based ordinal of the list item within its parent list.

Page
: (`page`) A reference to the current page.

PageInner
: (`page`) A reference to a page nested via the [`RenderShortcodes`] method. [See details](#pageinner-details).

Parent
: (`object`) The context of the parent list. See [Parent fields](#parent-fields) below.

Position
: (`string`) The position of the list item within the page content.

Text
: (`template.HTML`) The rendered list item content.

### Parent fields

IsOrdered
: (`bool`) Returns `true` if the parent list is an ordered list, or `false` if it is an unordered list.

## Examples

### Default rendering

To create render hooks that reproduce Hugo's default list output:

```go-html-template {file="layouts/_markup/render-list.html" copy=true}
{{ if .IsOrdered -}}
<ol{{ range $k, $v := .Attributes }} {{ $k }}="{{ $v }}"{{ end }}>
  {{ .Text }}
</ol>
{{- else -}}
<ul{{ range $k, $v := .Attributes }} {{ $k }}="{{ $v }}"{{ end }}>
  {{ .Text }}
</ul>
{{- end }}
```

```go-html-template {file="layouts/_markup/render-listitem.html" copy=true}
<li>{{ .Text }}</li>
```

### Semantic classes for recipe content

This example applies semantic CSS classes to list items based on whether the parent list is ordered (instructions) or unordered (ingredients):

```go-html-template {file="layouts/_markup/render-list.html" copy=true}
{{ if .IsOrdered }}
<ol{{ with .Attributes.class }} class="{{ . }}"{{ end }}>{{ .Text | safeHTML }}</ol>
{{ else }}
<ul{{ with .Attributes.class }} class="{{ . }}"{{ end }}>{{ .Text | safeHTML }}</ul>
{{ end }}
```

```go-html-template {file="layouts/_markup/render-listitem.html" copy=true}
<li class="{{ if .Parent.IsOrdered }}instruction{{ else }}ingredient{{ end }}">{{ .Text | safeHTML }}</li>
```

Given this Markdown:

```text
- 2 eggs
- 1 cup flour
- 1 cup milk

1. Mix the dry ingredients.
2. Add the wet ingredients.
3. Cook until golden.
```

The rendered output would be:

```html
<ul>
<li class="ingredient">2 eggs</li>
<li class="ingredient">1 cup flour</li>
<li class="ingredient">1 cup milk</li>
</ul>

<ol>
<li class="instruction">Mix the dry ingredients.</li>
<li class="instruction">Add the wet ingredients.</li>
<li class="instruction">Cook until golden.</li>
</ol>
```

{{% include "/_common/render-hooks/pageinner.md" %}}
