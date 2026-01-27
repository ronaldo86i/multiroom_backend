# Documentación del Servicio: Dispositivo Service

El `Dispositivo Service` actúa como el puente entre el software de gestión y el hardware físico (consolas, PCs, terminales). Es el encargado de mantener la comunicación bidireccional en tiempo real.

## 1. Responsabilidades
- Mantener conexiones WebSocket activas con los dispositivos físicos.
- Gestionar el registro y estado de dispositivos (`Activo`, `Inactivo`, `En uso`).
- Recibir comandos desde el `Sucursal Service` vía RabbitMQ (ej. "Encender Dispositivo X").
- Notificar cambios de estado a los clientes conectados.
- Gestión básica de clientes (datos personales y fidelización).

## 2. Tecnologías Específicas
- **WebSockets**: Utiliza `gofiber/contrib/websocket` para streaming bidireccional.
- **RabbitMQ**: Suscripción a exchanges/colas para recibir comandos de operación de otros servicios.
- **Doble Puerto**:
    - `HTTP (:8081)`: Para operaciones CRUD y consultas REST.
    - `WS (:8082)`: Dedicado exclusivamente al tráfico de WebSockets.

## 3. Listado Completo de Endpoints

### API HTTP (`/api/v1`)

#### Gestión de Dispositivos (`/dispositivos`)
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/dispositivos` | `dispositivo:ver` | Lista todos los dispositivos y su estado. |
| `GET` | `/dispositivos/byDispositivoId/:dispositivoId` | Token de Usuario | Busca dispositivo por su ID de hardware único. |
| `POST` | `/dispositivos` | Token de Usuario | Registra un nuevo dispositivo físico. |
| `PATCH` | `/dispositivos/:dispositivoId/habilitar` | `dispositivo:editar` | Activa un dispositivo para su uso. |
| `PATCH` | `/dispositivos/:dispositivoId/deshabilitar` | `dispositivo:editar` | Desactiva un dispositivo (mantenimiento/baja). |
| `DELETE` | `/dispositivos/:dispositivoId/eliminar` | `dispositivo:eliminar` | Elimina lógicamente un dispositivo. |

#### Consultas de Usuario (`/usuarios`)
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/usuarios/:usuarioId/dispositivos` | Token de Usuario | Lista dispositivos asignados a un usuario específico. |

#### Gestión de Clientes (`/clientes`)
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/clientes` | `cliente:ver` | Lista la base de datos de clientes. |
| `GET` | `/clientes/:clienteId` | `cliente:ver` | Detalle específico de cliente. |
| `POST` | `/clientes` | `cliente:crear` | Registra nuevo cliente (CRM). |
| `PUT` | `/clientes/:clienteId` | `cliente:editar` | Modifica datos del cliente. |
| `PATCH` | `/clientes/:clienteId/habilitar` | `cliente:editar` | Rehabilita a un cliente. |
| `PATCH` | `/clientes/:clienteId/deshabilitar` | `cliente:editar` | Bloquea/Banea a un cliente. |
| `DELETE` | `/clientes/:clienteId` | `cliente:eliminar` | Eliminación lógica de cliente. |

### API WebSocket (`/ws/v1`)
 Puerto dedicado para evitar bloqueo del hilo principal HTTP.
 
| Endpoint | Protocolo | Descripción |
| :--- | :--- | :--- |
| `/dispositivos/usuario/me` | `WSS` | Canal bidireccional principal. El dispositivo se autentica y mantiene la conexión viva para recibir comandos `START`, `PAUSE`, `STOP`. |

## 4. Flujo de Control de Hardware
1. El `Sucursal Service` procesa un inicio de tiempo.
2. El `Sucursal Service` publica un mensaje en RabbitMQ: `{ action: "START", target: "DISP_001" }`.
3. El `Dispositivo Service` recibe el mensaje.
4. Identifica la conexión WebSocket activa asociada a `DISP_001`.
5. Envía el comando binario o JSON por el socket para que el hardware se active.

## 5. Base de Datos
- `dispositivo`: Tabla central con el `dispositivo_id` único y estado de conexión.
- `cliente`: Información de contacto y fecha de nacimiento.
