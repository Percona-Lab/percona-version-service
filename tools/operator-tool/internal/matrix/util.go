package matrix

import (
	"reflect"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

func Iterate(matrix *vsAPI.VersionMatrix, f func(fieldName string, fieldValue reflect.Value) error) error {
	matrixType := reflect.TypeOf(matrix).Elem()
	matrixValue := reflect.ValueOf(matrix).Elem()

	for i := 0; i < matrixValue.NumField(); i++ {
		field := matrixType.Field(i)
		// check if value is exported
		if field.PkgPath != "" {
			continue
		}
		if err := f(field.Name, matrixValue.Field(i)); err != nil {
			return err
		}
	}
	return nil
}
