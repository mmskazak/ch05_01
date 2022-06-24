package main

//импорт пакетов
import (
	//работа с ошибками
	"errors"
	//блокировки памяти Mutex
	"sync"
)

// структура хранилища
var store = struct {
	//усовершенствованный Mutex для записи и чтения
	sync.RWMutex

	//мапа объявление
	m map[string]string
}{
	//мапа инициализация
	m: make(map[string]string),
}

// ErrorNoSuchKey инициализация ошибки, тип опущен
var ErrorNoSuchKey = errors.New("no such key")

// Delete функция Удаления экспортируемая и видна во ввсем пакете
func Delete(key string) error {
	//блокировка памяти
	store.Lock()
	//удаление
	delete(store.m, key)
	//разблокировака памяти
	store.Unlock()

	return nil
}

// Get функция получения значения из хранилища
func Get(key string) (string, error) {
	//блокировака возможно только чтение
	store.RLock()
	//порлучение значения из стороджа
	value, ok := store.m[key]
	store.RUnlock()

	if !ok {
		//если значния нет возвращаяем ошибку
		return "", ErrorNoSuchKey
	}

	return value, nil
}

// Put сохранеие данных в хранилище
func Put(key string, value string) error {
	//блокировка памяти
	store.Lock()
	//сохрание значения
	store.m[key] = value
	//разблоккировка памяти
	store.Unlock()

	return nil
}
