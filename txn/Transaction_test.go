package txn

import (
	"github.com/stretchr/testify/assert"
	"snapshot-isolation/mvcc"
	"testing"
)

func TestGetsANonExistingKeyInAReadonlyTransaction(t *testing.T) {
	memTable := mvcc.NewMemTable(10)

	transaction := NewReadonlyTransaction(memTable, NewOracle())
	_, ok := transaction.Get([]byte("non-existing"))

	assert.Equal(t, false, ok)
}

func TestGetsAnExistingKeyInAReadonlyTransaction(t *testing.T) {
	memTable := mvcc.NewMemTable(10)
	memTable.PutOrUpdate(mvcc.NewVersionedKey([]byte("HDD"), 1), mvcc.NewValue([]byte("Hard disk")))

	transaction := NewReadonlyTransaction(memTable, NewOracle())
	value, ok := transaction.Get([]byte("HDD"))

	assert.Equal(t, true, ok)
	assert.Equal(t, []byte("Hard disk"), value.Slice())
}

func TestCommitsAnEmptyReadWriteTransaction(t *testing.T) {
	memTable := mvcc.NewMemTable(10)

	oracle := NewOracle()
	transaction := NewReadWriteTransaction(memTable, oracle)

	_, err := transaction.Commit()

	assert.Error(t, err)
	assert.Equal(t, EmptyTransactionErr, err)
}

func TestGetsAnExistingKeyInAReadWriteTransaction(t *testing.T) {
	memTable := mvcc.NewMemTable(10)

	oracle := NewOracle()
	transaction := NewReadWriteTransaction(memTable, oracle)
	transaction.PutOrUpdate([]byte("HDD"), []byte("Hard disk"))
	transaction.PutOrUpdate([]byte("SSD"), []byte("Solid state disk"))

	done, _ := transaction.Commit()
	<-done

	readonlyTransaction := NewReadonlyTransaction(memTable, oracle)

	value, ok := readonlyTransaction.Get([]byte("HDD"))
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte("Hard disk"), value.Slice())

	value, ok = readonlyTransaction.Get([]byte("SSD"))
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte("Solid state disk"), value.Slice())

	_, ok = readonlyTransaction.Get([]byte("non-existing"))
	assert.Equal(t, false, ok)
}

func TestGetsTheValueFromAKeyInAReadWriteTransactionFromBatch(t *testing.T) {
	memTable := mvcc.NewMemTable(10)

	transaction := NewReadWriteTransaction(memTable, NewOracle())
	transaction.PutOrUpdate([]byte("HDD"), []byte("Hard disk"))

	value, ok := transaction.Get([]byte("HDD"))
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte("Hard disk"), value.Slice())

	done, _ := transaction.Commit()
	<-done
}

func TestTracksReadsInAReadWriteTransaction(t *testing.T) {
	memTable := mvcc.NewMemTable(10)

	transaction := NewReadWriteTransaction(memTable, NewOracle())
	transaction.PutOrUpdate([]byte("HDD"), []byte("Hard disk"))
	transaction.Get([]byte("SSD"))

	done, _ := transaction.Commit()
	<-done

	assert.Equal(t, 1, len(transaction.reads))
	key := transaction.reads[0]

	assert.Equal(t, []byte("SSD"), key)
}

func TestDoesNotTrackReadsInAReadWriteTransactionIfKeysAreReadFromTheBatch(t *testing.T) {
	memTable := mvcc.NewMemTable(10)

	transaction := NewReadWriteTransaction(memTable, NewOracle())
	transaction.PutOrUpdate([]byte("HDD"), []byte("Hard disk"))
	transaction.Get([]byte("HDD"))

	done, _ := transaction.Commit()
	<-done

	assert.Equal(t, 0, len(transaction.reads))
}
