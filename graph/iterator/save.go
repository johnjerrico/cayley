package iterator

import (
	"context"
	"fmt"

	"github.com/cayleygraph/cayley/graph"
)

var (
	_ graph.IteratorFuture = (*Save)(nil)
	_ graph.Tagger         = (*Save)(nil)
)

func Tag(it graph.Iterator, tag string) graph.Iterator {
	if s, ok := it.(graph.Tagger); ok {
		s.AddTags(tag)
		return s
	} else if s, ok := graph.As2(it).(graph.Tagger2); ok {
		s.AddTags(tag)
		return graph.AsLegacy(s)
	}
	return NewSave(it, tag)
}

func Tag2(it graph.Iterator2, tag string) graph.Iterator2 {
	if s, ok := it.(graph.Tagger2); ok {
		s.AddTags(tag)
		return s
	} else if s, ok := graph.AsLegacy(it).(graph.Tagger); ok {
		s.AddTags(tag)
		return graph.As2(s)
	}
	return newSave(it, tag)
}

func NewSave(on graph.Iterator, tags ...string) *Save {
	it := &Save{
		it: newSave(graph.As2(on), tags...),
	}
	it.Iterator = graph.NewLegacy(it.it)
	return it
}

type Save struct {
	it *save
	graph.Iterator
}

func (it *Save) As2() graph.Iterator2 {
	it.Close()
	return it.it
}

// Add a tag to the iterator.
func (it *Save) AddTags(tag ...string) {
	it.it.AddTags(tag...)
}

func (it *Save) AddFixedTag(tag string, value graph.Ref) {
	it.it.AddFixedTag(tag, value)
}

// Tags returns the tags held in the tagger. The returned value must not be mutated.
func (it *Save) Tags() []string {
	return it.it.Tags()
}

// Fixed returns the fixed tags held in the tagger. The returned value must not be mutated.
func (it *Save) FixedTags() map[string]graph.Ref {
	return it.it.FixedTags()
}

func (it *Save) CopyFromTagger(st graph.TaggerBase) {
	it.it.CopyFromTagger(st)
}

var (
	_ graph.Iterator2Compat = (*save)(nil)
	_ graph.Tagger2         = (*save)(nil)
)

func newSave(on graph.Iterator2, tags ...string) *save {
	s := &save{it: on}
	s.AddTags(tags...)
	return s
}

type save struct {
	it        graph.Iterator2
	tags      []string
	fixedTags map[string]graph.Ref
}

func (it *save) Iterate() graph.Iterator2Next {
	return newSaveNext(it.it.Iterate(), it.tags, it.fixedTags)
}

func (it *save) Lookup() graph.Iterator2Contains {
	return newSaveContains(it.it.Lookup(), it.tags, it.fixedTags)
}

func (it *save) AsLegacy() graph.Iterator {
	it2 := &Save{it: it}
	it2.Iterator = graph.NewLegacy(it)
	return it2
}

func (it *save) String() string {
	return fmt.Sprintf("Save(%v, %v)", it.tags, it.fixedTags)
}

// Add a tag to the iterator.
func (it *save) AddTags(tag ...string) {
	it.tags = append(it.tags, tag...)
}

func (it *save) AddFixedTag(tag string, value graph.Ref) {
	if it.fixedTags == nil {
		it.fixedTags = make(map[string]graph.Ref)
	}
	it.fixedTags[tag] = value
}

// Tags returns the tags held in the tagger. The returned value must not be mutated.
func (it *save) Tags() []string {
	return it.tags
}

// Fixed returns the fixed tags held in the tagger. The returned value must not be mutated.
func (it *save) FixedTags() map[string]graph.Ref {
	return it.fixedTags
}

func (it *save) CopyFromTagger(st graph.TaggerBase) {
	it.tags = append(it.tags, st.Tags()...)

	fixed := st.FixedTags()
	if len(fixed) == 0 {
		return
	}
	if it.fixedTags == nil {
		it.fixedTags = make(map[string]graph.Ref, len(fixed))
	}
	for k, v := range fixed {
		it.fixedTags[k] = v
	}
}

func (it *save) Stats() graph.IteratorStats {
	return it.it.Stats()
}

func (it *save) Size() (int64, bool) {
	return it.it.Size()
}

func (it *save) Optimize() (nit graph.Iterator2, no bool) {
	sub, ok := it.it.Optimize()
	if len(it.tags) == 0 && len(it.fixedTags) == 0 {
		return sub, true
	}
	if st, ok2 := sub.(graph.Tagger2); ok2 {
		st.CopyFromTagger(it)
		return st, true
	} else if st, ok2 := graph.AsLegacy(sub).(graph.Tagger); ok2 {
		st.CopyFromTagger(it)
		return graph.As2(st), true
	}
	if !ok {
		return it, false
	}
	s := newSave(sub)
	s.CopyFromTagger(it)
	return s, true
}

func (it *save) SubIterators() []graph.Iterator2 {
	return []graph.Iterator2{it.it}
}

func newSaveNext(it graph.Iterator2Next, tags []string, fixed map[string]graph.Ref) *saveNext {
	return &saveNext{it: it, tags: tags, fixedTags: fixed}
}

type saveNext struct {
	it        graph.Iterator2Next
	tags      []string
	fixedTags map[string]graph.Ref
}

func (it *saveNext) String() string {
	return fmt.Sprintf("Save(%v, %v)", it.tags, it.fixedTags)
}

func (it *saveNext) TagResults(dst map[string]graph.Ref) {
	it.it.TagResults(dst)

	v := it.Result()
	for _, tag := range it.tags {
		dst[tag] = v
	}

	for tag, value := range it.fixedTags {
		dst[tag] = value
	}
}

func (it *saveNext) Result() graph.Ref {
	return it.it.Result()
}

func (it *saveNext) Next(ctx context.Context) bool {
	return it.it.Next(ctx)
}

func (it *saveNext) NextPath(ctx context.Context) bool {
	return it.it.NextPath(ctx)
}

func (it *saveNext) Err() error {
	return it.it.Err()
}

func (it *saveNext) Close() error {
	return it.it.Close()
}

func newSaveContains(it graph.Iterator2Contains, tags []string, fixed map[string]graph.Ref) *saveContains {
	return &saveContains{it: it, tags: tags, fixed: fixed}
}

type saveContains struct {
	it    graph.Iterator2Contains
	tags  []string
	fixed map[string]graph.Ref
}

func (it *saveContains) String() string {
	return fmt.Sprintf("SaveContains(%v, %v)", it.tags, it.fixed)
}

func (it *saveContains) TagResults(dst map[string]graph.Ref) {
	it.it.TagResults(dst)

	v := it.Result()
	for _, tag := range it.tags {
		dst[tag] = v
	}

	for tag, value := range it.fixed {
		dst[tag] = value
	}
}

func (it *saveContains) Result() graph.Ref {
	return it.it.Result()
}

func (it *saveContains) NextPath(ctx context.Context) bool {
	return it.it.NextPath(ctx)
}

func (it *saveContains) Contains(ctx context.Context, v graph.Ref) bool {
	return it.it.Contains(ctx, v)
}

func (it *saveContains) Err() error {
	return it.it.Err()
}

func (it *saveContains) Close() error {
	return it.it.Close()
}
