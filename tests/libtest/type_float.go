package libtest

import (
	"database/sql"

	"testing"
)

// DoTestFloat tests the handling of the Float.
func DoTestFloat(t *testing.T) {
	TestForEachDB("TestFloat", t, testFloat)
	//
}

func testFloat(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesFloat))
	mySamples := make([]float64, len(samplesFloat))

	for i, sample := range samplesFloat {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, err := SetupTableInsert(db, tableName, "float", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv float64
	for rows.Next() {
		err = rows.Scan(&recv)
		if err != nil {
			t.Errorf("Scan failed on %dth scan: %v", i, err)
			continue
		}

		if recv != mySamples[i] {

			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", mySamples[i])
			t.Errorf("Received: %v", recv)
		}

		i++
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Error preparing rows: %v", err)
	}
}
