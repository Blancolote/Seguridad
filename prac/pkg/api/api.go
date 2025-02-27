// El paquete api contiene las estructuras necesarias
// para la comunicación entre servidor y cliente.
package api

const (
	ActionRegister           = "register"
	ActionLogin              = "login"
	ActionFetchData          = "fetchData"
	ActionUpdateData         = "updateData"
	ActionLogout             = "logout"
	ActionDarAlta            = "darAlta"
	ActionObtenerHistoriales = "obtenerHistoriales"
)

// Request y Response como antes
type Request struct { //omitempty es para que no aparezca en el json del request cuando se haga esta acción
	Action        string `json:"action"`
	Username      string `json:"username"`
	Password      string `json:"password,omitempty"`
	Token         string `json:"token,omitempty"`
	Data          string `json:"data,omitempty"`
	Especialidad  int    `json:"especialidad,omitempty"`
	Hospital      int    `json:"hospital,omitempty"`
	Apellido      string `json:"apellido,omitempty"`
	Sexo          string `json:"sexo,omitempty"`
	Nombre        string `json:"nombre,omitempty"`
	Nacimiento    string `json:"nacimiento;omitempty"` //COMPROBAR SI EXISTE TIPO FECHA Y SI NO HACER FUNCIÓN PARA COMPROBAR QUE ES UNA FECHA
	Medico        string `json:"medico;omitempty"`
	Observaciones string `json:"observaciones;omitempty"`
}

type Response struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Token     string `json:"token,omitempty"`
	Data      string `json:"data,omitempty"`
	Pacientes []int  `json:"pacientes,omitempty"` //lista con el id de los pacientes que tienen algún historial con su médico
}
