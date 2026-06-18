package nextengage

// Attribute operators mutate a contact attribute server-side (atomically)
// instead of replacing it. Use one as an attribute VALUE:
//
//	Attributes: map[string]any{
//	    "visits": nextengage.Increment(1),
//	    "tags":   nextengage.ArrayUnion("vip"),
//	}
//
// To delete an attribute key, set its value to nil — it marshals to JSON null,
// which the server treats as "remove this key".

// IncrementOp atomically adds Inc to a numeric attribute. Build it with
// Increment.
type IncrementOp struct {
	Inc float64 `json:"$inc"`
}

// ArrayOp adds and/or removes items from an array attribute. Build it with
// ArrayUnion or ArrayRemove.
type ArrayOp struct {
	Add    []any `json:"$add,omitempty"`
	Remove []any `json:"$remove,omitempty"`
}

// Increment atomically adds n to a numeric attribute (negative to decrement). A
// missing or non-number attribute starts from 0.
//
//	Attributes: map[string]any{"visits": nextengage.Increment(1)}
func Increment(n float64) IncrementOp {
	return IncrementOp{Inc: n}
}

// ArrayUnion adds items to an array attribute (deduped server-side).
//
//	Attributes: map[string]any{"tags": nextengage.ArrayUnion("pro", "vip")}
func ArrayUnion(items ...any) ArrayOp {
	return ArrayOp{Add: items}
}

// ArrayRemove removes items from an array attribute.
//
//	Attributes: map[string]any{"tags": nextengage.ArrayRemove("beta")}
func ArrayRemove(items ...any) ArrayOp {
	return ArrayOp{Remove: items}
}
