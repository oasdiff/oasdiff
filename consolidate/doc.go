// Package consolidate merges findings that report one change into a single
// finding. The checker reports a change wherever it applies; each mode here
// recognizes one way a change ends up reported several times, and findings a
// mode does not recognize pass through unchanged.
package consolidate
