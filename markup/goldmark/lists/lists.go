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

package lists

import (
	"fmt"

	"github.com/gohugoio/hugo/common/herrors"
	"github.com/gohugoio/hugo/common/types/hstring"
	"github.com/gohugoio/hugo/markup/converter/hooks"
	"github.com/gohugoio/hugo/markup/goldmark/internal/render"
	"github.com/gohugoio/hugo/markup/internal/attributes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type (
	listsExtension struct{}
	htmlRenderer   struct{}
)

// New returns a goldmark Extender that enables the render-list and render-listitem hooks.
func New() goldmark.Extender {
	return &listsExtension{}
}

func (e *listsExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newHTMLRenderer(), 100),
	))
}

func newHTMLRenderer() renderer.NodeRenderer {
	return &htmlRenderer{}
}

func (r *htmlRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
}

func (r *htmlRenderer) renderList(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	ctx := w.(*render.Context)
	n := node.(*ast.List)

	if entering {
		ctx.PushPos(ctx.Buffer.Len())
		// Push a per-list item ordinal counter so renderListItem can track
		// the zero-based index of each item within this list.
		ctx.PushValue(ast.KindListItem, new(int))
		return ast.WalkContinue, nil
	}

	// Clean up the item ordinal counter before processing list output.
	ctx.PopValue(ast.KindListItem)

	text := ctx.PopRenderedString()
	ordinal := ctx.GetAndIncrementOrdinal(ast.KindList)

	renderer := ctx.RenderContext().GetRenderer(hooks.ListRendererType, nil)
	if renderer == nil {
		return r.renderListDefault(w, n, text)
	}

	lctx := &listContext{
		BaseContext:      render.NewBaseContext(ctx, renderer, n, src, nil, ordinal),
		text:             hstring.HTML(text),
		isOrdered:        n.IsOrdered(),
		AttributesHolder: attributes.New(n.Attributes(), attributes.AttributesOwnerGeneral),
	}

	cr := renderer.(hooks.ListRenderer)
	err := cr.RenderList(ctx.RenderContext().Ctx, w, lctx)
	if err != nil {
		return ast.WalkContinue, herrors.NewFileErrorFromPos(err, lctx.Position())
	}

	return ast.WalkContinue, nil
}

func (r *htmlRenderer) renderListItem(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	ctx := w.(*render.Context)
	n := node.(*ast.ListItem)

	if entering {
		ctx.PushPos(ctx.Buffer.Len())
		return ast.WalkContinue, nil
	}

	// Consume the per-list ordinal counter (pushed by renderList entering).
	ordinal := 0
	if ptr := ctx.PeekValue(ast.KindListItem); ptr != nil {
		p := ptr.(*int)
		ordinal = *p
		*p++
	}

	text := ctx.PopRenderedString()

	renderer := ctx.RenderContext().GetRenderer(hooks.ListItemRendererType, nil)
	if renderer == nil {
		return r.renderListItemDefault(w, n, text)
	}

	parent := n.Parent().(*ast.List)
	lictx := &listItemContext{
		BaseContext: render.NewBaseContext(ctx, renderer, n, src, nil, ordinal),
		text:        hstring.HTML(text),
		parent:      &listItemParentContext{isOrdered: parent.IsOrdered()},
	}

	cr := renderer.(hooks.ListItemRenderer)
	err := cr.RenderListItem(ctx.RenderContext().Ctx, w, lictx)
	if err != nil {
		return ast.WalkContinue, herrors.NewFileErrorFromPos(err, lictx.Position())
	}

	return ast.WalkContinue, nil
}

// renderListDefault reproduces goldmark's default list rendering, used as fallback.
func (r *htmlRenderer) renderListDefault(w util.BufWriter, n *ast.List, text string) (ast.WalkStatus, error) {
	tag := "ul"
	if n.IsOrdered() {
		tag = "ol"
	}
	_ = w.WriteByte('<')
	_, _ = w.WriteString(tag)
	if n.IsOrdered() && n.Start != 1 {
		_, _ = fmt.Fprintf(w, " start=\"%d\"", n.Start)
	}
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ListAttributeFilter)
	}
	_, _ = w.WriteString(">\n")
	_, _ = w.WriteString(text)
	_, _ = w.WriteString("</")
	_, _ = w.WriteString(tag)
	_, _ = w.WriteString(">\n")
	return ast.WalkContinue, nil
}

// renderListItemDefault reproduces goldmark's default list item rendering, used as fallback.
func (r *htmlRenderer) renderListItemDefault(w util.BufWriter, n *ast.ListItem, text string) (ast.WalkStatus, error) {
	if n.Attributes() != nil {
		_, _ = w.WriteString("<li")
		html.RenderAttributes(w, n, html.ListItemAttributeFilter)
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("<li>")
	}
	// In loose lists the first child is a Paragraph (not TextBlock), and goldmark
	// emits a newline after <li> in that case.
	if fc := n.FirstChild(); fc != nil {
		if _, ok := fc.(*ast.TextBlock); !ok {
			_ = w.WriteByte('\n')
		}
	}
	_, _ = w.WriteString(text)
	_, _ = w.WriteString("</li>\n")
	return ast.WalkContinue, nil
}

type listContext struct {
	hooks.BaseContext
	text      hstring.HTML
	isOrdered bool
	*attributes.AttributesHolder
}

func (c *listContext) Text() hstring.HTML { return c.text }
func (c *listContext) IsOrdered() bool    { return c.isOrdered }

type listItemContext struct {
	hooks.BaseContext
	text   hstring.HTML
	parent *listItemParentContext
}

func (c *listItemContext) Text() hstring.HTML                  { return c.text }
func (c *listItemContext) Parent() hooks.ListItemParentContext { return c.parent }

type listItemParentContext struct {
	isOrdered bool
}

func (c *listItemParentContext) IsOrdered() bool { return c.isOrdered }
