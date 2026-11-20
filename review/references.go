package review

import (
	"reflect"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
)

// A change to an endpoint can be made in a component the endpoint refers to.
// The change is then located in the component, and what ties it to the
// endpoint is the $ref in the endpoint that leads there, directly or through
// other components. refGraph finds that $ref.

// lineRange is a run of lines in one file.
type lineRange struct {
	file       string
	start, end int
}

func (r lineRange) contains(file string, line int) bool {
	return r.file == file && r.start <= line && line <= r.end
}

func (r lineRange) intersects(other lineRange) bool {
	return r.file == other.file && r.start <= other.end && other.start <= r.end
}

// refSite is one $ref: where it is written, and the lines of what it refers to.
type refSite struct {
	file   string
	line   int
	target lineRange
}

// refGraph holds every $ref of a set of documents.
type refGraph struct {
	sites    []refSite           // ordered by file, then line
	byTarget map[lineRange][]int // the sites referring to each target
	targets  map[string][]lineRange
	leading  map[lineRange]map[int]bool
}

func newRefGraph(docs ...*openapi3.T) *refGraph {
	g := &refGraph{
		byTarget: map[lineRange][]int{},
		targets:  map[string][]lineRange{},
		leading:  map[lineRange]map[int]bool{},
	}
	visited := map[uintptr]bool{}
	for _, doc := range docs {
		if doc != nil {
			walkRefs(reflect.ValueOf(doc), visited, func(site refSite) { g.sites = append(g.sites, site) })
		}
	}
	sort.SliceStable(g.sites, func(i, j int) bool {
		if g.sites[i].file != g.sites[j].file {
			return g.sites[i].file < g.sites[j].file
		}
		return g.sites[i].line < g.sites[j].line
	})
	for i, site := range g.sites {
		if _, seen := g.byTarget[site.target]; !seen {
			g.targets[site.target.file] = append(g.targets[site.target.file], site.target)
		}
		g.byTarget[site.target] = append(g.byTarget[site.target], i)
	}
	for _, targets := range g.targets {
		sort.Slice(targets, func(i, j int) bool { return targets[i].start < targets[j].start })
	}
	return g
}

// walkRefs visits every $ref below v, following each one into what it refers
// to, so the $refs of a referenced file are visited too.
func walkRefs(v reflect.Value, visited map[uintptr]bool, emit func(refSite)) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		// These keep their entries in an unexported map.
		switch x := v.Interface().(type) {
		case *openapi3.Origin:
			return
		case *openapi3.Paths:
			walkRefs(reflect.ValueOf(x.Map()), visited, emit)
			return
		case *openapi3.Responses:
			walkRefs(reflect.ValueOf(x.Map()), visited, emit)
			return
		case *openapi3.Callback:
			walkRefs(reflect.ValueOf(x.Map()), visited, emit)
			return
		}
		if visited[v.Pointer()] {
			return
		}
		visited[v.Pointer()] = true
		walkRefs(v.Elem(), visited, emit)
	case reflect.Struct:
		if ref, value := v.FieldByName("Ref"), v.FieldByName("Value"); ref.IsValid() && ref.Kind() == reflect.String && value.IsValid() && value.Kind() == reflect.Pointer {
			if ref.String() != "" {
				if site, ok := newRefSite(v, value); ok {
					emit(site)
				}
			}
			walkRefs(value, visited, emit)
			return
		}
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				walkRefs(v.Field(i), visited, emit)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			walkRefs(v.Index(i), visited, emit)
		}
	case reflect.Map:
		for iter := v.MapRange(); iter.Next(); {
			walkRefs(iter.Value(), visited, emit)
		}
	}
}

// newRefSite reads a $ref's own location and the lines of what it refers to,
// and reports whether both are known.
func newRefSite(ref, value reflect.Value) (refSite, bool) {
	at := originOf(ref)
	if at == nil || value.IsNil() {
		return refSite{}, false
	}
	written, ok := at.Fields.Lookup("$ref")
	if !ok || written.Line == 0 {
		return refSite{}, false
	}
	target := originOf(value.Elem())
	start, end, ok := originEndRange(target)
	if !ok {
		return refSite{}, false
	}
	return refSite{
		file:   written.File,
		line:   written.Line,
		target: lineRange{file: target.Key.File, start: start, end: end},
	}, true
}

func originOf(v reflect.Value) *openapi3.Origin {
	field := v.FieldByName("Origin")
	if !field.IsValid() {
		return nil
	}
	origin, _ := field.Interface().(*openapi3.Origin)
	return origin
}

// container returns the innermost target a line is written in: the component
// a $ref on that line belongs to.
func (g *refGraph) container(file string, line int) (lineRange, bool) {
	targets := g.targets[file]
	// Targets are ordered by start, so the last one starting at or before the
	// line that also reaches it is the innermost.
	for i := sort.Search(len(targets), func(i int) bool { return targets[i].start > line }) - 1; i >= 0; i-- {
		if targets[i].end >= line {
			return targets[i], true
		}
	}
	return lineRange{}, false
}

// leadingTo returns the sites that lead to the given lines: the $refs to a
// target overlapping them, and the $refs to whatever those are written in.
func (g *refGraph) leadingTo(lines lineRange) map[int]bool {
	if sites, ok := g.leading[lines]; ok {
		return sites
	}
	sites := map[int]bool{}
	seen := map[lineRange]bool{}
	var queue []lineRange
	for _, target := range g.targets[lines.file] {
		if target.intersects(lines) {
			seen[target] = true
			queue = append(queue, target)
		}
	}
	for len(queue) > 0 {
		target := queue[0]
		queue = queue[1:]
		for _, i := range g.byTarget[target] {
			sites[i] = true
			if outer, ok := g.container(g.sites[i].file, g.sites[i].line); ok && !seen[outer] {
				seen[outer] = true
				queue = append(queue, outer)
			}
		}
	}
	g.leading[lines] = sites
	return sites
}

// refLine returns the line of the first $ref written in from that leads to
// to, preferring one on a line prefer accepts, or 0 when there is none.
func (g *refGraph) refLine(from, to lineRange, prefer func(line int) bool) int {
	leading := g.leadingTo(to)
	first := sort.Search(len(g.sites), func(i int) bool {
		site := g.sites[i]
		return site.file > from.file || site.file == from.file && site.line >= from.start
	})
	fallback := 0
	for i := first; i < len(g.sites) && from.contains(g.sites[i].file, g.sites[i].line); i++ {
		if !leading[i] {
			continue
		}
		if prefer(g.sites[i].line) {
			return g.sites[i].line
		}
		if fallback == 0 {
			fallback = g.sites[i].line
		}
	}
	return fallback
}

// referenceFinder finds, for a change located outside its endpoint, the
// endpoint's block and the $ref in it that leads to the change.
type referenceFinder struct {
	baseDocs, revDocs []*openapi3.T
	baseIdx, revIdx   docIndex
	baseRefs, revRefs *refGraph
	partOf            map[string]endpointPart
}

func newReferenceFinder(baseDocs, revDocs []*openapi3.T, baseIdx, revIdx docIndex) *referenceFinder {
	return &referenceFinder{baseDocs: baseDocs, revDocs: revDocs, baseIdx: baseIdx, revIdx: revIdx}
}

// find returns the endpoint's block and the reference to record in it, for a
// change that names an endpoint and resolved to another block.
func (f *referenceFinder) find(c checker.Change, located resolution) (resolution, Reference, bool) {
	operation, path := c.GetOperation(), c.GetPath()
	if operation == "" || path == "" {
		return resolution{}, Reference{}, false
	}
	key := operation + " " + path
	if located.title == key {
		return resolution{}, Reference{}, false
	}

	file := firstFile(located.rev, located.base)
	endpoint := newResolution(key, f.baseIdx.spanFor(key, file), f.revIdx.spanFor(key, file), f.baseIdx, f.revIdx)

	// The graphs are built on the first change that needs them: most reviews
	// have no change located outside its endpoint.
	if f.baseRefs == nil {
		f.baseRefs, f.revRefs = newRefGraph(f.baseDocs...), newRefGraph(f.revDocs...)
		f.partOf = partsByRule()
	}
	part := f.partOf[c.GetId()]
	reference := Reference{
		Fingerprint: checker.Fingerprint(c),
		BaseLine:    refLine(f.baseRefs, f.baseDocs, endpoint.base, located.base, operation, path, part),
		RevLine:     refLine(f.revRefs, f.revDocs, endpoint.rev, located.rev, operation, path, part),
	}
	if reference.BaseLine == 0 && reference.RevLine == 0 {
		return resolution{}, Reference{}, false
	}
	return endpoint, reference, true
}

func firstFile(spans ...*span) string {
	for _, s := range spans {
		if s != nil {
			return s.file
		}
	}
	return ""
}

// endpointPart is the part of an endpoint a rule is about.
type endpointPart int

const (
	anyPart endpointPart = iota
	parametersPart
	requestBodyPart
	responsesPart
)

// refLine returns the line of the $ref in the endpoint that leads to where
// the change is located, on one side. It prefers a $ref in the part of the
// endpoint the change is about, since an endpoint often refers to the same
// component from its parameters, its request body and its responses.
func refLine(refs *refGraph, docs []*openapi3.T, endpoint, located *span, method, path string, part endpointPart) int {
	if endpoint == nil || located == nil {
		return 0
	}
	from := lineRange{file: endpoint.file, start: endpoint.start, end: endpoint.end}
	to := lineRange{file: located.file, start: located.start, end: located.end}

	operation := operationAt(docs, from, method, path)
	return refs.refLine(from, to, func(line int) bool { return part == anyPart || partAt(operation, line) == part })
}

// operationAt returns the operation written at the given lines.
func operationAt(docs []*openapi3.T, endpoint lineRange, method, path string) *openapi3.Operation {
	for _, doc := range docs {
		if doc == nil || doc.Paths == nil {
			continue
		}
		pathItem := doc.Paths.Value(path)
		if pathItem == nil {
			continue
		}
		operation := pathItem.GetOperation(method)
		if operation == nil || operation.Origin == nil || operation.Origin.Key == nil {
			continue
		}
		if endpoint.contains(operation.Origin.Key.File, operation.Origin.Key.Line) {
			return operation
		}
	}
	return nil
}

// partAt returns the part of the operation a line is written in. A line
// outside the request body and the responses is taken as the parameters',
// the only other part of an operation whose $refs lead to a schema.
func partAt(operation *openapi3.Operation, line int) endpointPart {
	if operation == nil {
		return anyPart
	}
	within := func(start, end int, ok bool) bool { return ok && start <= line && line <= end }
	if operation.Responses != nil && within(originEndRange(operation.Responses.Origin)) {
		return responsesPart
	}
	if within(requestBodyRange(operation.RequestBody)) {
		return requestBodyPart
	}
	return parametersPart
}

// requestBodyRange returns the lines of an operation's request body: the
// $ref line when it is a reference, the body's own lines otherwise.
func requestBodyRange(requestBody *openapi3.RequestBodyRef) (start, end int, ok bool) {
	if requestBody == nil {
		return 0, 0, false
	}
	if requestBody.Ref != "" {
		if requestBody.Origin == nil {
			return 0, 0, false
		}
		written, ok := requestBody.Origin.Fields.Lookup("$ref")
		return written.Line, written.Line, ok && written.Line != 0
	}
	if requestBody.Value == nil {
		return 0, 0, false
	}
	return originEndRange(requestBody.Value.Origin)
}

// partsByRule returns, by rule id, the part of an endpoint a rule is about.
func partsByRule() map[string]endpointPart {
	rules := checker.GetAllRules()
	m := make(map[string]endpointPart, len(rules))
	for _, r := range rules {
		switch {
		case r.Direction == checker.DirectionResponse:
			m[r.Id] = responsesPart
		case r.Area == checker.AreaParameters:
			m[r.Id] = parametersPart
		case r.Direction == checker.DirectionRequest:
			m[r.Id] = requestBodyPart
		}
	}
	return m
}
