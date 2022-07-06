package reflect

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	assert "github.com/smartystreets/goconvey/convey"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func Test_ReflectPathSelect(t *testing.T) {

	assert.Convey("Test_ReflectPathSelect", t, func() {

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

		assert.Convey("single struct field", func() {
			assert.Convey("normal", func() {
				selection, ok := PathSelect(ctx, "String", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, "string!")
			})
			assert.Convey("nested", func() {
				selection, ok := PathSelect(ctx, "Nested.Pointer.Pointer.Map.a", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, 1)
			})
		})

		assert.Convey("single struct json tag", func() {
			assert.Convey("normal json tag string", func() {
				selection, ok := PathSelect(ctx, "string", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, "string!")
			})
			assert.Convey("custom json tag int", func() {
				selection, ok := PathSelect(ctx, "json-int", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, 101)
			})
			assert.Convey("Nested json tag", func() {
				selection, ok := PathSelect(ctx, "nested.pointer.pointer.map.a", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, 1)
			})
		})

		assert.Convey("hybrid struct field AND json tag", func() {
			selection, ok := PathSelect(ctx, "Nested.pointer.Pointer.map.a", foo)
			assert.So(ok, assert.ShouldBeTrue)
			assert.So(selection, assert.ShouldResemble, 1)
		})

		assert.Convey("with map", func() {
			assert.Convey("simple map", func() {
				selection, ok := PathSelect(ctx, "Nested.map.a", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, 1)
			})
			assert.Convey("complex map", func() {
				selection, ok := PathSelect(ctx, "nested.MapI.pointer.pointer.map.a", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, 1)
			})
		})

		assert.Convey("with slice", func() {
			selection, ok := PathSelect(ctx, "nested.Slice.1", foo)
			assert.So(ok, assert.ShouldBeTrue)
			assert.So(selection, assert.ShouldResemble, 2)
		})

		assert.Convey("with Nil", func() {
			selection, ok := PathSelect(ctx, "string", nil)
			fmt.Println(selection, ok)
			assert.So(ok, assert.ShouldBeFalse)
			assert.So(selection, assert.ShouldResemble, nil)
		})

		assert.Convey("should not panic", func() {
			assert.So(func() {
				PathSelect(ctx, "string", nil)
				PathSelect(ctx, "string", foo.Nil)
				PathSelect(ctx, "", nil)
				PathSelect(ctx, "", foo.Nil)
				PathSelect(ctx, "unexported", foo)
			}, assert.ShouldNotPanic)
		})

		assert.Convey("should false", func() {
			_, o1 := PathSelect(ctx, "string", nil)
			assert.So(o1, assert.ShouldBeFalse)

			_, o2 := PathSelect(ctx, "string", foo.Nil)
			assert.So(o2, assert.ShouldBeFalse)

			_, o3 := PathSelect(ctx, "", nil)
			assert.So(o3, assert.ShouldBeFalse)

			_, o4 := PathSelect(ctx, "", foo.Nil)
			assert.So(o4, assert.ShouldBeFalse)

			_, o5 := PathSelect(ctx, "unexported", foo)
			assert.So(o5, assert.ShouldBeFalse)
		})

		assert.Convey("with shadow", func() {
			assert.Convey("struct", func() {
				selection, ok := PathSelect(ctx, "shadow_name", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, "shadow!")
			})

			assert.Convey("pointer", func() {
				selection, ok := PathSelect(ctx, "nested.shadow_name", foo)
				assert.So(ok, assert.ShouldBeTrue)
				assert.So(selection, assert.ShouldResemble, "shadow pointer!")
			})
		})

	})
}
