// El paquete store provee una interfaz genérica de almacenamiento.
// Cada motor (en nuestro caso bbolt) se implementa en un archivo separado
// que debe cumplir la interfaz Store.
package store

import "fmt"

// Store define los métodos comunes que deben implementar
// los diferentes motores de almacenamiento.
type Store interface {
	// Put almacena (o actualiza) el valor 'value' bajo la clave 'key'
	// dentro del 'namespace' indicado.
	Put(namespace string, key, value []byte) error

	// Get recupera el valor asociado a la clave 'key'
	// dentro del 'namespace' especificado.
	Get(namespace string, key []byte) ([]byte, error)

	// Delete elimina la clave 'key' dentro del 'namespace' especificado.
	Delete(namespace string, key []byte) error

	// ListKeys devuelve todas las claves existentes en el namespace.
	ListKeys(namespace string) ([][]byte, error)

	// KeysByPrefix devuelve las claves que empiecen con 'prefix' dentro
	// del namespace especificado.
	KeysByPrefix(namespace string, prefix []byte) ([][]byte, error)

	// Close cierra cualquier recurso abierto (por ej. cerrar la base de datos).
	Close() error

	// Dump imprime todo el contenido de la base de datos para depuración de errores.
	Dump() error
}

// NewStore permite instanciar diferentes tipos de Store
// dependiendo del motor solicitado (sólo se soporta "bbolt").
func NewStore(engine, path string) (Store, error) {
	switch engine {
	case "bbolt":
		return NewBboltStore(path)
	default:
		return nil, fmt.Errorf("motor de almacenamiento desconocido: %s", engine)
	}
}

func main() {
	// Crear una instancia de Store usando bbolt
	db, err := NewStore("bbolt", "mi_base.db") // Reemplaza con la ruta real de la DB
	if err != nil {
		fmt.Printf("Error al abrir la base de datos: %v", err)
	}
	defer db.Close()

	// Imprimir todo el contenido de la base de datos
	fmt.Println("Contenido de la base de datos:")
	if err := db.Dump(); err != nil {
		fmt.Printf("Error al hacer dump de la base de datos: %v", err)
	}
}
