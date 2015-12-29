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

package data

import "sync"

// Go rutine safe structuer for storing data.

type dataMap map[string]interface{}

type Data struct {
	data dataMap
	lock *sync.Mutex
}

func NewData() *Data {
	return &Data{
		data: make(dataMap),
		lock: new(sync.Mutex),
	}
}

func (d *Data) Set(key string, value interface{}) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.data[key] = value
}

func (d *Data) Get(key string, def ...interface{}) (interface{}, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	value, ok := d.data[key]
	if !ok && def != nil {
		return def, false
	} else {
		return value, ok
	}
}

func (d *Data) Snapshot() dataMap {
	d.lock.Lock()
	defer d.lock.Unlock()
	newData := make(dataMap)
	for k, v := range d.data {
		newData[k] = v
	}
	return newData
}
