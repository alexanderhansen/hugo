// Copyright 2025 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lists_test

import (
	"testing"

	"github.com/gohugoio/hugo/hugolib"
)

func TestListHook(t *testing.T) {
	t.Parallel()
	files := `
-- hugo.toml --
[markup.goldmark.parser.attribute]
block = true
-- content/p1.md --
---
title: "p1"
---

- alpha
- beta

1. first
2. second

- sole item
{.myclass}

-- layouts/single.html --
{{ .Content }}
-- layouts/_markup/render-list.html --
<LIST ordered={{ .IsOrdered }} ordinal={{ .Ordinal }} attrs="{{ .Attributes }}" page="{{ .Page.Title }}">{{ .Text | safeHTML }}</LIST>
`
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/p1/index.html",
		// unordered list is ordinal 0
		`<LIST ordered=false ordinal=0`,
		// ordered list is ordinal 1
		`<LIST ordered=true ordinal=1`,
		// third list (unordered with attribute) is ordinal 2
		`<LIST ordered=false ordinal=2`,
		// attributes passed through
		`attrs="map[class:myclass]"`,
		// page context populated
		`page="p1"`,
		// text contains default-rendered list items
		`<li>alpha</li>`,
		`<li>first</li>`,
	)
}

func TestListItemHook(t *testing.T) {
	t.Parallel()
	files := `
-- hugo.toml --
-- content/p1.md --
---
title: "p1"
---

- apple
- banana
- cherry

1. one
2. two

-- layouts/single.html --
{{ .Content }}
-- layouts/_markup/render-listitem.html --
<LI ordered={{ .Parent.IsOrdered }} ordinal={{ .Ordinal }} page="{{ .Page.Title }}">{{ .Text | safeHTML }}</LI>
`
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/p1/index.html",
		// unordered items have ordered=false and correct ordinals
		`<LI ordered=false ordinal=0 page="p1">apple`,
		`<LI ordered=false ordinal=1 page="p1">banana`,
		`<LI ordered=false ordinal=2 page="p1">cherry`,
		// ordered items have ordered=true and ordinals reset per list
		`<LI ordered=true ordinal=0 page="p1">one`,
		`<LI ordered=true ordinal=1 page="p1">two`,
	)
}

func TestListBothHooks(t *testing.T) {
	t.Parallel()
	files := `
-- hugo.toml --
-- content/p1.md --
---
title: "p1"
---

- x
- y

1. a
2. b

-- layouts/single.html --
{{ .Content }}
-- layouts/_markup/render-list.html --
<LIST ordered={{ .IsOrdered }}>{{ .Text | safeHTML }}</LIST>
-- layouts/_markup/render-listitem.html --
<LI ordered={{ .Parent.IsOrdered }}>{{ .Text | safeHTML }}</LI>
`
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/p1/index.html",
		// list hook wraps listitem hook output
		"<LIST ordered=false><LI ordered=false>x</LI>\n<LI ordered=false>y</LI>\n</LIST>",
		"<LIST ordered=true><LI ordered=true>a</LI>\n<LI ordered=true>b</LI>\n</LIST>",
	)
}

func TestListDefault(t *testing.T) {
	t.Parallel()
	files := `
-- hugo.toml --
[markup.goldmark.parser.attribute]
block = true
-- content/p1.md --
---
title: "p1"
---

- alpha
- beta

1. first
2. second

- classed
{.mylist}

-- layouts/single.html --
{{ .Content }}
`
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/p1/index.html",
		// unordered list — default goldmark output
		"<ul>\n<li>alpha</li>\n<li>beta</li>\n</ul>",
		// ordered list
		"<ol>\n<li>first</li>\n<li>second</li>\n</ol>",
		// list with attribute
		`<ul class="mylist">`,
	)
}

func TestListDefaultNested(t *testing.T) {
	t.Parallel()
	files := `
-- hugo.toml --
-- content/p1.md --
---
title: "p1"
---

- outer 1
  - inner a
  - inner b
- outer 2

-- layouts/single.html --
{{ .Content }}
`
	b := hugolib.Test(t, files)

	// Nested lists render correctly without hooks.
	b.AssertFileContent("public/p1/index.html",
		"<ul>",
		"<li>outer 1",
		"<li>inner a</li>",
		"<li>inner b</li>",
		"<li>outer 2</li>",
	)
}

func TestListItemHookNestedList(t *testing.T) {
	t.Parallel()
	files := `
-- hugo.toml --
-- content/p1.md --
---
title: "p1"
---

- outer 1
  - inner a
  - inner b
- outer 2

-- layouts/single.html --
{{ .Content }}
-- layouts/_markup/render-listitem.html --
<LI ordered={{ .Parent.IsOrdered }} ordinal={{ .Ordinal }}>{{ .Text | safeHTML }}</LI>
`
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/p1/index.html",
		// inner items ordinals are independent from outer
		`<LI ordered=false ordinal=0>inner a`,
		`<LI ordered=false ordinal=1>inner b`,
		// outer items
		`<LI ordered=false ordinal=0>outer 1`,
		`<LI ordered=false ordinal=1>outer 2`,
	)
}
