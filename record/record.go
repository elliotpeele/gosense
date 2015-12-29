/*
 * Copyright (c) Elliot Peele <elliot@bentlogic.net>
 *
 * This program is distributed under the terms of the MIT License as found
 * in a file called LICENSE. If it is not present, the license
 * is always available at http://www.opensource.org/licenses/mit-license.php.
 *
 * This program is distributed in the hope that it will be useful, but
 * without any warrenty; without even the implied warranty of merchantability
 * or fitness for a particular purpose. See the MIT License for full details.
 */

package record

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	Application string    `json:"application"`
	Hostname    string    `json:"hostname"`
	Name        string    `json:"name"`
	DataType    string    `json:"type"`
	Extra       string    `json:"extra"`
	Value       float64   `json:"value"`
	TimeStamp   time.Time `json:"time"`
}

func ParseRecord(in string) (*Record, error) {
	parts := strings.Split(in, " ")
	kparts := strings.Split(parts[0], ".")
	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}
	tparts := strings.Split(parts[2], ".")
	sec, err := strconv.ParseInt(tparts[0], 10, 64)
	if err != nil {
		return nil, err
	}
	nsec, err := strconv.ParseInt(tparts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	return &Record{
		Application: kparts[0],
		Hostname:    kparts[1],
		Name:        kparts[2],
		DataType:    kparts[3],
		Extra:       strings.Join(kparts[4:], "."),
		Value:       value,
		TimeStamp:   time.Unix(sec, nsec),
	}, nil
}

func (r *Record) Key() string {
	return strings.Join([]string{
		r.Application,
		r.Hostname,
		r.Name,
		r.DataType,
		r.Extra,
	}, ".")
}

func (r *Record) JSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(r); err != nil {
		return err
	}
	return nil
}
