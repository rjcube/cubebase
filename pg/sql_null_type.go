package pg

import "database/sql"

func NullStringToString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func NullTimeToString(nt sql.NullTime) *string {
	if nt.Valid {
		t := &nt.Time
		s := t.Format("2006-01-02 15:04:05")
		return &s
	}
	return nil
}

func NullTimeToTimestampString(nt sql.NullTime) *string {
	if nt.Valid {
		t := &nt.Time
		s := t.Format("2006-01-02 15:04:05")
		return &s
	}
	return nil
}

func NullTimeToDateString(nt sql.NullTime) *string {
	if nt.Valid {
		t := &nt.Time
		s := t.Format("2006-01-02")
		return &s
	}
	return nil
}

func NullTimeToFormatString(nt sql.NullTime, format string) *string {
	if nt.Valid {
		t := &nt.Time
		s := t.Format(format)
		return &s
	}
	return nil
}
