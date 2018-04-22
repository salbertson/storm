package v3

import (
	"errors"

	"github.com/asdine/storm/v3/engine"
)

func selectt(e engine.Engine, pl engine.Pipeline, path ...string) (engine.Bucket, error) {
	tx, err := e.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.Bucket(path...)
	if err != nil {
		return nil, err
	}

	b, err = pl.Run(b)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func limit(l int) engine.Pipe {
	return func(b engine.Bucket) (engine.Bucket, error) {
		schema, err := b.Schema()
		if err != nil {
			return nil, err
		}

		buff := engine.NewRecordBuffer(schema)

		for i := 0; i < l; i++ {
			r, err := b.Next()
			if err != nil {
				return nil, err
			}

			if r == nil {
				break
			}

			buff.Add(r)
		}

		return buff, nil
	}
}

func maxInt64(field string) engine.Pipe {
	return func(b engine.Bucket) (engine.Bucket, error) {
		schema, err := b.Schema()
		if err != nil {
			return nil, err
		}

		var max int64

		var scanner engine.RecordScanner

		f, ok := schema.Fields[field]
		if !ok {
			return nil, errors.New("field not found")
		}

		if f.Type != engine.Int64Field {
			return nil, errors.New("field incompatible with max")
		}

		for {
			r, err := b.Next()
			if err != nil {
				return nil, err
			}

			if r == nil {
				break
			}

			scanner.Record = r
			i, err := scanner.GetInt64(field)
			if err != nil {
				return nil, err
			}

			if i > max {
				max = i
			}
		}

		newField := "max(" + field + ")"

		// ugly
		buff := engine.NewRecordBuffer(&engine.Schema{
			Fields: map[string]*engine.Field{
				newField: &engine.Field{
					Name: newField,
					Type: engine.Int64Field,
				}},
		})

		var fb engine.FieldBuffer

		fb.Add(&engine.Field{
			Name:  newField,
			Type:  engine.Int64Field,
			Value: max,
		})

		buff.Add(&fb)

		return buff, nil
	}
}
