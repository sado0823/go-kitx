package errorx

import "bytes"

type (
	// A Batch is an error that can hold multiple errorx.
	Batch struct {
		errs errorArray
	}

	errorArray []error
)

// Add adds errs to be, nil errorx are ignored.
func (be *Batch) Add(errs ...error) {
	for _, err := range errs {
		if err != nil {
			be.errs = append(be.errs, err)
		}
	}
}

// Err returns an error that represents all errorx.
func (be *Batch) Err() error {
	switch len(be.errs) {
	case 0:
		return nil
	case 1:
		return be.errs[0]
	default:
		return be.errs
	}
}

// NotNil checks if any error inside.
func (be *Batch) NotNil() bool {
	return len(be.errs) > 0
}

// Error returns a string that represents inside errorx.
func (ea errorArray) Error() string {
	var buf bytes.Buffer

	for i := range ea {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(ea[i].Error())
	}

	return buf.String()
}
