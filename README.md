Permite consultar patentes de autos a partir del rut del dueño. 
Consulta el api del registro civil de chile

### Ejemplo de uso

```go
package main

import (
	"fmt"
	"github.com/Eitol/patentes_cl"
)

func main() {
	client := patentes_cl.NewClient()
	rut := "27029012" // Sustituye esto con un RUT real
	vehicles, err := client.GetByRut(rut)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Vehicles:", vehicles)
}
```
 
### ¿Como funciona internamente?

Consististe en dos pasos

Paso 1: Generar un token
```http
POST /PortalRvm/oauth/token HTTP/1.1
Host: pagosrvm.srcei.cl
Authorization: Basic ${CREDENTIALS}

grant_type=password&username=PORT_RVM_2021&password=E5ED00A617CADBBCE1C11EF3689FCFDD2CF599E1CA9B71D0F8E2CDD592F593DD
```

Generará un token como este:

```json
{
"access_token": "eyJhbGc....",
}
```

Paso 2: Usar ese token para hacer la busqueda
```http
GET /PortalRvm/api/lista/ppu/${RUT} HTTP/1.1
Host: pagosrvm.srcei.cl
Content-type: application/json
Authorization: bearer ${TOKEN GENERADO EN EL PASO 1}
```

La respuesta será como esta:
```json
[
  {
    "ppu": "RTPD60",
    "marca": "CHEVROLET",
    "modelo": "TRACKER 1.2T",
    "tipo": "STATION WAGON",
    "aFabricacion": "",
    "nroMotor": "",
    "nroChasis": "",
    "nroSerie": "",
    "nroVin": "",
    "codigoColorBase": "",
    "descColorBase": "",
    "restoColor": "",
    "calidad": "0",
    "dvPpu": "9",
    "tipoPropietario": "N"
  }
]
```

Y cuando el token esté expirado será como esta
```json
{"error":"invalid_token","error_description":"Access token expired:blablabla"}
```