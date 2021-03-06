package libtest

import (
	"database/sql"

	"testing"
)

// DoTestBit tests the handling of the Bit.
func DoTestBit(t *testing.T) {
	TestForEachDB("TestBit", t, testBit)
	//
}

func testBit(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesBit))
	mySamples := make([]bool, len(samplesBit))

	for i, sample := range samplesBit {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, err := SetupTableInsert(db, tableName, "bit", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv bool
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
