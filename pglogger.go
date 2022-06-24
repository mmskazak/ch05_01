package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"sync"
)

type PostgresDbParam struct {
	dbName   string
	host     string
	user     string
	password string
}

type PostgresTransactionLogger struct {
	events chan<- Event //Write-only channel for sending events
	errors <-chan error //Read-only chanel for receiving errors
	db     *sql.DB      // Our database access
	wg     *sync.WaitGroup
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.wg.Add(1)
	l.events <- Event{EventType: EventPut, Key: key, Value: url.QueryEscape(value)}
}

func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.wg.Add(1)
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *PostgresTransactionLogger) Lastsequence() uint64 {
	return 0
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16) //Make events chanel
	l.events = events

	errors := make(chan error, 1) //Make an errors channel
	l.errors = errors

	go func() { //The INSERT` query
		query := `INSERT INTO transactions (event_type,key, value) VALUES ($1,$2,$3)`

		for e := range events { //Retrieve the next Event
			_, err := l.db.Exec( //Execute the INSERT query
				query,
				e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
			}
		}
	}()
}

func (l *PostgresTransactionLogger) Wait() {
	l.wg.Wait()
}

func (l *PostgresTransactionLogger) Close() error {
	l.wg.Wait()

	if l.events != nil {
		close(l.events)
	}

	return l.db.Close()
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)    // An unbuffered events channel
	outError := make(chan error, 1) // A buffered error

	query := "SELECT sequence, event_type, key, value FROM transactions"

	go func() {
		defer close(outEvent)
		defer close(outError)

		rows, err := l.db.Query(query) // Run query; get result set
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}

		defer rows.Close() // This is important!

		var e Event // Create an empty Event

		for rows.Next() { //Iterate over the rows

			err = rows.Scan( // Read the values
				&e.Sequence, &e.EventType, //row into the Event
				&e.Key, &e.Value)

			if err != nil {
				outError <- err
				return
			}

			outEvent <- e //Send event to channel
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()

	return outEvent, outError
}

func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	const table = "transactions"

	var result string

	rows, err := l.db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", table))
	defer rows.Close()

	if err != nil {
		return false, err
	}

	for rows.Next() && result != table {
		rows.Scan(&result)
	}

	return result == table, rows.Err()
}

//func(l *PostgresTransactionLogger) createTable() error {
//	var err error
//
//	 createQuery := `CREATE TABLE transactions (
//			sequence BIGSERIAL PRIMARY KEY,
//			event_type SMALLINT
//		)`
//}
