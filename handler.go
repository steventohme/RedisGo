package main

import (
	"sync"
	"strconv"
)

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"SET":  set,
	"GET":  get,
	"HSET": hset,
	"HGET": hget,
	"HGETALL": hgetall,
	"MGET": mget,
	"INCR": incr,
	"DECR": decr,
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}
var HSETs = map[string]map[string]string{}
var HSETsMu = sync.RWMutex{}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "Invalid number of arguments"}
	}
	key := args[0].bulk
	value := args[1].bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "Invalid number of arguments"}
	}
	key := args[0].bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "Invalid number of arguments"}
	}

	hash := args[0].bulk
	key := args[1].bulk
	value := args[2].bulk

	HSETsMu.Lock()
	_, ok := HSETs[hash]
	if !ok {
		HSETs[hash] = map[string]string{}
	}
	HSETs[hash][key] = value
	HSETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "Invalid number of arguments"}
	}

	hash := args[0].bulk
	key := args[1].bulk

	HSETsMu.RLock()
	value, ok := HSETs[hash][key]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hgetall(args []Value) Value {
	if len(args) != 1{
		return Value{typ: "error", str: "Invalid number of arguments"}
	}

	hash := args[0].bulk

	HSETsMu.RLock()
	values, ok := HSETs[hash]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	array := make([]Value, 0, len(values) * 2)
	for key, value := range values {
		array = append(array, Value{typ: "bulk", bulk: key})
		array = append(array, Value{typ: "bulk", bulk: value})
	}

	return Value{typ: "array", array: array}
}

func mget (args []Value) Value {
	if len(args) < 1 {
		return Value{typ: "error", str: "Invalid number of arguments"}
	}

	array := make([]Value, 0, len(args))
	for _, arg := range args {
		if arg.typ != "bulk" {
			return Value{typ: "error", str: "Invalid argument type"}
		}

		SETsMu.RLock()
		value, ok := SETs[arg.bulk]
		SETsMu.RUnlock()

		if !ok {
			array = append(array, Value{typ: "null"})
			continue
		}

		array = append(array, Value{typ: "bulk", bulk: value})
	}

	return Value{typ: "array", array: array}
}

func incr (args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "Invalid number of arguments"}
	}

	key := args[0].bulk

	SETsMu.Lock()
	value, ok := SETs[key]
	if !ok {
		SETs[key] = "1"
		SETsMu.Unlock()
		return Value{typ: "bulk", bulk: "1"}
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		SETsMu.Unlock()
		return Value{typ: "error", str: "Value cannot be converted to integer"}
	}

	i++
	value = strconv.Itoa(i)
	SETs[key] = value

	SETsMu.Unlock()
	

	return Value{typ: "bulk", bulk: value}
}

func decr (args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "Invalid number of arguments"}
	}

	key := args[0].bulk

	SETsMu.Lock()
	value, ok := SETs[key]
	if !ok {
		SETs[key] = "1"
		SETsMu.Unlock()
		return Value{typ: "bulk", bulk: "1"}
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		SETsMu.Unlock()
		return Value{typ: "error", str: "Value cannot be converted to integer"}
	}

	i--
	value = strconv.Itoa(i)
	SETs[key] = value

	SETsMu.Unlock()
	

	return Value{typ: "bulk", bulk: value}
}