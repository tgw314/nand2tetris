package symboltable

import "fmt"

type SymbolTable struct {
	table map[string]int
	tail  int
}

func New() *SymbolTable {
	return &SymbolTable{
		table: map[string]int{
			"R0":     0,
			"R1":     1,
			"R2":     2,
			"R3":     3,
			"R4":     4,
			"R5":     5,
			"R6":     6,
			"R7":     7,
			"R8":     8,
			"R9":     9,
			"R10":    10,
			"R11":    11,
			"R12":    12,
			"R13":    13,
			"R14":    14,
			"R15":    15,
			"SP":     0,
			"LCL":    1,
			"ARG":    2,
			"THIS":   3,
			"THAT":   4,
			"SCREEN": 16384,
			"KBD":    24576,
		},
		tail: 16,
	}
}

func (st *SymbolTable) AddEntry(symbol string, address int) {
	st.table[symbol] = address
}

func (st *SymbolTable) AddVar(symbol string) {
	st.table[symbol] = st.tail
	st.tail++
}

func (st *SymbolTable) Contains(symbol string) bool {
	_, ok := st.table[symbol]
	return ok
}

func (st *SymbolTable) GetAddress(symbol string) (int, error) {
	address, ok := st.table[symbol]
	if !ok {
		return -1, fmt.Errorf("symbol %s not found", symbol)
	}

	return address, nil
}
