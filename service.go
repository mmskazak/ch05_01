package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	//сервер
	"net/http"

	//роутинг от gorilla
	"github.com/gorilla/mux"
)

var transact *TransactionLogger

// loggingMiddleware обычный мидделвар
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//логирование все запросов в консоли
		log.Println(r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

//notAllowedHandler просто функция 403
func notAllowedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Allowed", http.StatusMethodNotAllowed)
}

// keyValuePutHandler метод сохраняет данные в хранилище
func keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	// получение массива аргуметов из реквеста
	vars := mux.Vars(r)
	//получение занчения ключа
	key := vars["key"]

	log.Println("key = ", key)

	// метод ридОлл читает тело целиком, реализует интерфейс ридер
	//поллучает занчение
	value, err := ioutil.ReadAll(r.Body)

	log.Println("value = ", value)

	//закрытие в конце процесса чтения тела реквеста
	defer r.Body.Close()

	if err != nil {
		//возвращаем ответ, пишем ишибку в него
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//сохраняем занчение в хранилище
	log.Println("//сохраняем занчение в хранилище")
	err = Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("//пишем в файл")
	transact.WritePut(key, string(value))

	//добавляем в ответ заголовок
	log.Println("//добавляем в ответ заголовок")
	w.WriteHeader(http.StatusCreated)

	//логируем в консоль состоявшеся событие
	log.Printf("PUT key = %s value = %s", key, string(value))
}

func keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	//получаем данные из реквеста через компонет gorilla/mux
	vars := mux.Vars(r)
	//получаем {key}
	key := vars["key"]

	//получаем значение
	value, err := Get(key)
	if errors.Is(err, ErrorNoSuchKey) {
		//пишем ошибку err.Error() == no such key 404
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//в респонс записываем найденные данные

	w.Write([]byte(value))

	log.Printf("GET key=%s\n", key)
}

// keyValueDeleteHandler удаляет данные из хранилища
func keyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// получение аргументов реквеста
	vars := mux.Vars(r)
	//получение значния аргумента key
	key := vars["key"]

	// удаление значения из хранилища
	err := Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func initializeTransactionLog() error {
	var err error

	transact, err = NewTransactionLogger("transactions.log")
	log.Println("transactions.log")
	log.Println(transact)
	if err != nil {
		fmt.Errorf("failed to create transaction logger: %w", err)
	}

	events, errs := transact.ReadEvents()
	count, ok, e := 0, true, Event{}

	for ok && err == nil {
		select {
		case err, ok = <-errs:
		case e, ok = <-events:
			switch e.EventType {
			case EventDelete: //Got a DELETE event!
				err = Delete(e.Key)
				count++
			case EventPut: //Got a PUT event
				err = Put(e.Key, e.Value)
				count++
			}
		}
	}

	log.Printf("%d events replayed \n", count)

	transact.Run()

	return err
}

func main() {

	err := initializeTransactionLog()
	if err != nil {
		panic(err)
	}
	log.Println("service work")

	r := mux.NewRouter()

	r.Use(loggingMiddleware)

	r.HandleFunc("/v1/{key}", keyValueGetHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", keyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", keyValueDeleteHandler).Methods("DELETE")

	r.HandleFunc("/v1", notAllowedHandler)
	r.HandleFunc("/v1/{key}", notAllowedHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
