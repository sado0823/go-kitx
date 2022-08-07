package reflectx

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func Test_ReflectPathSelect(t *testing.T) {

	type (
		shadow struct {
			ShadowName string `json:"shadow_name"`
		}
		dummyNestedParameter struct {
			*shadow
			Funk    string
			Map     map[string]int `json:"map"`
			Slice   []int
			MapI    map[string]interface{}
			Pointer interface{} `json:"pointer"`
		}
		dummyParameter struct {
			shadow
			unexported string
			String     string `json:"string"`
			Int        int    `json:"json-int"`
			BoolFalse  bool
			Nil        interface{}
			Map        map[string]int
			MapI       map[string]interface{}
			Nested     dummyNestedParameter `json:"nested"`
		}
	)

	var (
		ctx = context.Background()
		foo = dummyParameter{
			shadow:     shadow{ShadowName: "shadow!"},
			unexported: "unexported!",
			String:     "string!",
			Int:        101,
			BoolFalse:  false,
			Nil:        nil,
			Nested: dummyNestedParameter{
				shadow: &shadow{ShadowName: "shadow pointer!"},
				Funk:   "funk",
				Map:    map[string]int{"a": 1, "b": 2, "c": 3},
				Slice:  []int{1, 2, 3},
				MapI: map[string]interface{}{
					"pointer": &dummyNestedParameter{
						Funk:  "pointer!",
						Map:   map[string]int{"a": 1},
						Slice: []int{1, 2, 3},
						MapI:  nil,
						Pointer: &dummyNestedParameter{
							Funk:    "pointer",
							Map:     map[string]int{"a": 1},
							Slice:   []int{1, 2, 3},
							MapI:    nil,
							Pointer: nil,
						},
					},
				},
				Pointer: &dummyNestedParameter{
					Funk:  "pointer!",
					Map:   map[string]int{"a": 1},
					Slice: []int{1, 2, 3},
					MapI:  nil,
					Pointer: &dummyNestedParameter{
						Funk:    "pointer",
						Map:     map[string]int{"a": 1},
						Slice:   []int{1, 2, 3},
						MapI:    nil,
						Pointer: nil,
					},
				},
			},
		}
	)

	t.Run("single struct field", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "String", foo)
			assert.True(t, ok)
			assert.Exactly(t, "string!", selection)
		})
		t.Run("nested", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "Nested.Pointer.Pointer.Map.a", foo)
			assert.True(t, ok)
			assert.Exactly(t, 1, selection)
		})
	})

	t.Run("single struct json tag", func(t *testing.T) {
		t.Run("normal json tag string", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "string", foo)
			assert.True(t, ok)
			assert.Exactly(t, "string!", selection)
		})
		t.Run("custom json tag int", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "json-int", foo)
			assert.True(t, ok)
			assert.Exactly(t, 101, selection)
		})
		t.Run("Nested json tag", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "nested.pointer.pointer.map.a", foo)
			assert.True(t, ok)
			assert.Exactly(t, 1, selection)
		})
	})

	t.Run("hybrid struct field AND json tag", func(t *testing.T) {
		selection, ok := PathSelect(ctx, "Nested.pointer.Pointer.map.a", foo)
		assert.True(t, ok)
		assert.Exactly(t, 1, selection)
	})

	t.Run("with map", func(t *testing.T) {
		t.Run("simple map", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "Nested.map.a", foo)
			assert.True(t, ok)
			assert.Exactly(t, 1, selection)
		})
		t.Run("complex map", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "nested.MapI.pointer.pointer.map.a", foo)
			assert.True(t, ok)
			assert.Exactly(t, 1, selection)
		})
	})

	t.Run("with slice", func(t *testing.T) {
		selection, ok := PathSelect(ctx, "nested.Slice.1", foo)
		assert.True(t, ok)
		assert.Exactly(t, 2, selection)
	})

	t.Run("with Nil", func(t *testing.T) {
		selection, ok := PathSelect(ctx, "string", nil)
		fmt.Println(selection, ok)
		assert.False(t, ok)
		assert.Nil(t, selection)
	})

	t.Run("should not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			PathSelect(ctx, "string", nil)
			PathSelect(ctx, "string", foo.Nil)
			PathSelect(ctx, "", nil)
			PathSelect(ctx, "", foo.Nil)
			PathSelect(ctx, "unexported", foo)
		})
	})

	t.Run("should false", func(t *testing.T) {
		_, o1 := PathSelect(ctx, "string", nil)
		assert.False(t, o1)

		_, o2 := PathSelect(ctx, "string", foo.Nil)
		assert.False(t, o2)

		_, o3 := PathSelect(ctx, "", nil)
		assert.False(t, o3)

		_, o4 := PathSelect(ctx, "", foo.Nil)
		assert.False(t, o4)

		_, o5 := PathSelect(ctx, "unexported", foo)
		assert.False(t, o5)
	})

	t.Run("with shadow", func(t *testing.T) {
		t.Run("struct", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "shadow_name", foo)
			assert.True(t, ok)
			assert.Exactly(t, "shadow!", selection)
		})

		t.Run("pointer", func(t *testing.T) {
			selection, ok := PathSelect(ctx, "nested.shadow_name", foo)
			assert.True(t, ok)
			assert.Exactly(t, "shadow pointer!", selection)
		})
	})
}
