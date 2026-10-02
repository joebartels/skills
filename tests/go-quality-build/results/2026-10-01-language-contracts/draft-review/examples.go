package draftreview
import "errors"
func completionError(primary, finish error) error {
    if primary == nil {
        return finish
    }
    if finish == nil {
        return primary
    }
    return errors.Join(primary, finish) // import "errors"; Go 1.20+
}

type Problem struct{}
func (*Problem) Error() string { return "problem" }
func typedNil() error {
var concrete *Problem = nil
var err error = concrete // non-nil interface, if *Problem implements error

return err
}
