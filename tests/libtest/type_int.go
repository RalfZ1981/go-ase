package libtest

import (
	"database/sql"

	"testing"
)

// DoTestInt tests the handling of the Int.
func DoTestInt(t *testing.T) {
	TestForEachDB("TestInt", t, testInt)
	//
}

func testInt(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesInt))
	mySamples := make([]int32, len(samplesInt))

	for i, sample := range samplesInt {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, err := SetupTableInsert(db, tableName, "int", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv int32
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
