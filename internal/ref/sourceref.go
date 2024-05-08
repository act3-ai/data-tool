package ref

// SourceRef is an interface that provides a Ref object for a source, as well as a custom string that is intended to
// be an initial declaration.  This enables a Ref to be resolved from a declaration, but have the original information
// preserved for additional purposes, such as maintaining a record of processed references.
type SourceRef interface {
	// GetRef returns the translated and parsed Ref structure for a SourceRef
	GetRef() Ref

	// GetDecl returns a string representing the original declaration of a reference (before translation to Ref)
	GetDecl() string
}

// DeclRef is an extension of Ref that adds tracking for a declaration string.
type DeclRef struct {
	Ref
	decl string
}

// GetRef implements SourceRef interface, and returns the internal Ref object.
func (dr DeclRef) GetRef() Ref {
	return dr.Ref
}

// GetDecl implements SourceRef interface, and returns the stored declaration string.
func (dr DeclRef) GetDecl() string {
	return dr.decl
}

// NewDeclaredRef generates a SourceRef (internally, DeclRef) from a resolved bottle ref and an original declaration
// string.
func NewDeclaredRef(resolvedRef Ref, declaration string) SourceRef {
	return &DeclRef{resolvedRef, declaration}
}

// NewStringDeclaredRef returns a SourceRef (internally, DeclRef) based on the string form declaration of a Ref.
// This function assumes that the declaration can be parsed to a Ref.  If unsure, resolve the reference elsewhere
// and use NewDeclaredRef.
func NewStringDeclaredRef(declaration string) SourceRef {
	return &DeclRef{
		RepoFromString(declaration),
		declaration,
	}
}

// Set defines an interface for retrieving a list of refs from a source provider.  The Set is not a strict set,
// duplicates are possible and can be differentiated by retrieving the declaration matching the index of a given ref
// using the Decl function.
type Set interface {
	Refs() []SourceRef

	// RefByReg should return a map associating a list of SourceRef to a registry or other association
	RefByReg() map[string][]SourceRef
}
