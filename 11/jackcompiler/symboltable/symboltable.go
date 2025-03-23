package symboltable

type SymbolKind string

const (
	STATIC = SymbolKind("static")
	FIELD  = SymbolKind("field")
	ARG    = SymbolKind("arg")
	VAR    = SymbolKind("var")
	NONE   = SymbolKind("none")
)

type symbol struct {
	name  string
	ty    string
	kind  SymbolKind
	index int
}

type SymbolTable struct {
	staticCount int
	fieldCount  int
	argCount    int
	varCount    int
	table       []symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{}
}

func (st *SymbolTable) Reset() {
	*st = SymbolTable{}
}

func (st *SymbolTable) countPtr(kind SymbolKind) *int {
	switch kind {
	case STATIC:
		return &st.staticCount
	case FIELD:
		return &st.fieldCount
	case ARG:
		return &st.argCount
	case VAR:
		return &st.varCount
	}

	return nil
}

func (st *SymbolTable) Define(name string, ty string, kind SymbolKind) {
	p := st.countPtr(kind)
	st.table = append(st.table, symbol{name, ty, kind, *p})
	*p++
}

func (st *SymbolTable) VarCount(kind SymbolKind) int {
	return *st.countPtr(kind)
}

func (st *SymbolTable) KindOf(name string) SymbolKind {
	for _, s := range st.table {
		if s.name == name {
			return s.kind
		}
	}
	return NONE
}

func (st *SymbolTable) TypeOf(name string) string {
	for _, s := range st.table {
		if s.name == name {
			return s.ty
		}
	}
	return ""
}

func (st *SymbolTable) IndexOf(name string) int {
	for _, s := range st.table {
		if s.name == name {
			return s.index
		}
	}
	return -1
}
