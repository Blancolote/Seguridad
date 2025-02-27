// El paquete client contiene la lógica de interacción con el usuario
// así como de comunicación con el servidor.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"prac/pkg/api"
	"prac/pkg/ui"
)

// client estructura interna no exportada que controla
// el estado de la sesión (usuario, token) y logger.
type client struct {
	log          *log.Logger
	currentUser  string
	authToken    string
	apellido     string
	hospital     int
	especialidad int
}

// Run es la única función exportada de este paquete.
// Crea un client interno y ejecuta el bucle principal.
func Run() {
	// Creamos un logger con prefijo 'cli' para identificar
	// los mensajes en la consola.
	c := &client{
		log: log.New(os.Stdout, "[cli] ", log.LstdFlags),
	}
	c.runLoop()
}

// runLoop maneja la lógica del menú principal.
// Si NO hay usuario logueado, se muestran ciertas opciones;
// si SÍ hay usuario logueado, se muestran otras.
func (c *client) runLoop() {
	for {
		ui.ClearScreen()

		// Construimos un título que muestre el usuario logueado, si lo hubiera.
		var title string
		if c.currentUser == "" {
			title = "Menú"
		} else {
			title = fmt.Sprintf("Menú (%s)", c.currentUser)
		}

		// Generamos las opciones dinámicamente, según si hay un login activo.
		var options []string
		if c.currentUser == "" {
			// Usuario NO logueado: Registro, Login, Salir
			options = []string{
				"Registrar usuario",
				"Iniciar sesión",
				"Salir",
			}
		} else {
			// Usuario logueado: Ver datos, Actualizar datos, Logout, Salir
			options = []string{
				"Dar de alta paciente",
				"Ver información paciente",
				"Crear expediente",
				"Cerrar sesión",
				"Salir",
			}
		}

		// Mostramos el menú y obtenemos la elección del usuario.
		choice := ui.PrintMenu(title, options)

		// Hay que mapear la opción elegida según si está logueado o no.
		if c.currentUser == "" {
			// Caso NO logueado
			switch choice {
			case 1:
				c.registerUser()
			case 2:
				c.loginUser()
			case 3:
				// Opción Salir
				c.log.Println("Saliendo del cliente...")
				return
			}
		} else {
			// Caso logueado
			switch choice {
			case 1:
				c.fetchData()
			case 2:
				c.darAlta()
			case 3:
				c.logoutUser()
			case 4:
				// Opción Salir
				c.log.Println("Saliendo del cliente...")
				return
			}
		}

		// Pausa para que el usuario vea resultados.
		ui.Pause("Pulsa [Enter] para continuar...")
	}
}

// loginUser pide credenciales y realiza un login en el servidor.
func (c *client) loginUser() {
	ui.ClearScreen()
	fmt.Println("** Inicio de sesión **")

	username := ui.ReadInput("Nombre de usuario")
	password := ui.ReadInput("Contraseña")

	res := c.sendRequest(api.Request{
		Action:   api.ActionLogin,
		Username: username,
		Password: password,
	})

	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)

	// Si login fue exitoso, guardamos currentUser y el token.
	if res.Success {
		c.currentUser = username
		c.authToken = res.Token
		fmt.Println("Sesión iniciada con éxito. Token guardado.")
	}
}

// registerUser pide credenciales y las envía al servidor para un registro.
// Si el registro es exitoso, se intenta el login automático.
func (c *client) registerUser() {

	ui.ClearScreen()
	fmt.Println("** Registro de usuario **")

	username := ui.ReadInput("Nombre de usuario")
	password := ui.ReadInput("Contraseña")
	especialidad := ui.ReadInt("Especialidad (1:Obstetricia;2:Oncología;3:Traumatología;0 para salir)")
	hospital := ui.ReadInt("Hospital (0:San Juan de Alicante;1:Doctor Balmis;2:Hospital de Elda)")
	apellido := ui.ReadInput("Apellido")

	for { //ESTE BUCLE ES PARA QUE HASTA QUE NO PONGA BIEN LA ESPECIALIDAD NO SALGA SI NO PULSA s
		if especialidad == 1 || especialidad == 2 || especialidad == 3 {
			break
		}
		if especialidad == 0 {
			//AQUÍ SALIR AL MENÚ DE INICIO
		}
		fmt.Println("Error del valor añadido. Inténtelo de nuevo o pulse 0 para salir")
		especialidad = ui.ReadInt("Especialidad (01:Obstetricia;2:Oncología;3:Traumatología;0 para salir)")
	}

	for { //ESTE BUCLE ES PARA QUE HASTA QUE NO PONGA BIEN EL HOSPITAL NO SALGA SI NO PULSA s
		if hospital == 1 || hospital == 2 || hospital == 3 {
			break
		}
		if hospital == 0 {
			//AQUÍ SALIR AL MENÚ DE INICIO
		}
		fmt.Println("Error del valor añadido. Inténtelo de nuevo o pulse 0 para salir")
		hospital = ui.ReadInt("Especialidad (1:Obstetricia;2:Oncología;3:Traumatología;0 para salir)")
	}

	// Enviamos la acción al servidor
	res := c.sendRequest(api.Request{
		Action:       api.ActionRegister,
		Username:     username,
		Password:     password,
		Especialidad: especialidad,
		Hospital:     hospital,
		Apellido:     apellido,
	})

	// Mostramos resultado
	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)

	// Si fue exitoso, probamos loguear automáticamente.
	if res.Success {
		c.log.Println("Registro exitoso; intentando login automático...")

		loginRes := c.sendRequest(api.Request{
			Action:   api.ActionLogin,
			Username: username,
			Password: password,
		})
		if loginRes.Success {
			c.currentUser = username
			c.authToken = loginRes.Token
			fmt.Println("Login automático exitoso. Token guardado.")
		} else {
			fmt.Println("No se ha podido hacer login automático:", loginRes.Message)
		}
	}
}

// fetchData pide datos privados al servidor.
// El servidor devuelve la data asociada al usuario logueado.
func (c *client) fetchData() {
	ui.ClearScreen()
	fmt.Println("** Obtener datos del usuario **")

	// Chequeo básico de que haya sesión
	if c.currentUser == "" || c.authToken == "" {
		fmt.Println("No estás logueado. Inicia sesión primero.")
		return
	}

	// Hacemos la request con ActionFetchData
	res := c.sendRequest(api.Request{
		Action:   api.ActionFetchData,
		Username: c.currentUser,
		Token:    c.authToken,
	})

	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)

	// Si fue exitoso, mostramos la data recibida
	if res.Success {
		fmt.Println("Tus datos:", res.Data)
	}
}

// dar alta a un paciente que no existe RECORDAR QUE PUEDE SER QUE EL PACIENTE SÍ QUE EXISTE, HACER QUE EL SERVIDOR LO COMPRUEBE
func (c *client) darAlta() {
	ui.ClearScreen()
	fmt.Println("** Actualizar datos del usuario **")

	if c.currentUser == "" || c.authToken == "" {
		fmt.Println("No estás logueado. Inicia sesión primero.")
		return
	}

	nombre := ui.ReadInput("Nombre del paciente")
	primerApellido := ui.ReadInput("Primer apellido del paciente")
	nacimiento := ui.ReadInput("Fecha de nacimiento del paciente")
	sexo := ui.ReadInput("Sexo del paciente (M:mujer;H:hombre)")

	// Enviamos la solicitud de actualización
	res := c.sendRequest(api.Request{
		Action:     api.ActionUpdateData,
		Username:   c.currentUser,
		Token:      c.authToken,
		Nombre:     nombre,
		Apellido:   primerApellido,
		Nacimiento: nacimiento,
		Hospital:   c.hospital,
		Sexo:       sexo,
	})

	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)
}

func (c *client) verInfo() { //se piden los historiales que estén relacionados con el médico (username)
	ui.ClearScreen()
	fmt.Println("** Buscador de pacientes **")

	if c.currentUser == "" || c.authToken == "" {
		fmt.Println("No estás logueado. Inicia sesión primero.")
		return
	}

	res := c.sendRequest(api.Request{
		Action:   api.ActionObtenerHistoriales,
		Username: c.currentUser,
		Token:    c.authToken,
	})

	for i := 0; i < 5; i++ {
		paciente := res.Pacientes[i] //supuestamente es una lista con datos del paciente

		fmt.Println("------------------------------------------")
		fmt.Println("paciente.nombre", "paciente.apellido", " || ", "paciente.fechaCreacion")
	}
}

// logoutUser llama a la acción logout en el servidor, y si es exitosa,
// borra la sesión local (currentUser/authToken).
func (c *client) logoutUser() {
	ui.ClearScreen()
	fmt.Println("** Cerrar sesión **")

	if c.currentUser == "" || c.authToken == "" {
		fmt.Println("No estás logueado.")
		return
	}

	// Llamamos al servidor con la acción ActionLogout
	res := c.sendRequest(api.Request{
		Action:   api.ActionLogout,
		Username: c.currentUser,
		Token:    c.authToken,
	})

	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)

	// Si fue exitoso, limpiamos la sesión local.
	if res.Success {
		c.currentUser = ""
		c.authToken = ""
	}
}

// sendRequest envía un POST JSON a la URL del servidor y
// devuelve la respuesta decodificada. Se usa para todas las acciones.
func (c *client) sendRequest(req api.Request) api.Response {
	jsonData, _ := json.Marshal(req)
	resp, err := http.Post("http://localhost:8080/api", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error al contactar con el servidor:", err)
		return api.Response{Success: false, Message: "Error de conexión"}
	}
	defer resp.Body.Close()

	// Leemos el body de respuesta y lo desempaquetamos en un api.Response
	body, _ := io.ReadAll(resp.Body)
	var res api.Response
	_ = json.Unmarshal(body, &res)
	return res
}
