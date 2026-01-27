# Documentación del Servicio: Sucursal Service

El `Sucursal Service` es el corazón operativo y financiero del sistema Multiroom. Es el servicio más extenso y gestiona desde la geografía del negocio hasta el control de inventario y ventas.

## 1. Responsabilidades
- **Administración Geográfica**: Gestión de países y sucursales.
- **Operación de Salas**: Control de tiempos de uso, pausas, reanudaciones y cancelaciones.
- **Logística e Inventario**: Compras a proveedores, transferencias entre sucursales y ajustes de stock.
- **Ventas (POS)**: Generación de ventas, cobros múltiples y anulación de tickets.
- **Reportes**: Generación de PDFs para comprobantes y cierre de caja.

## 2. Tecnologías Específicas
- **RabbitMQ**: Publicación de comandos hacia el `Dispositivo Service`.
- **Maroto v2**: Librería para la generación técnica de documentos PDF (facturas/reportes).
- **Public Assets**: Gestión de archivos públicos (estáticos) como fotos de productos y archivos de configuración regional.

## 3. Listado Completo de Endpoints (`/api/v1`)

### Organización Geográfica
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/paises` | `pais:ver` | Lista países soportados. |
| `GET` | `/paises/:paisId` | `pais:ver` | Detalle país. |
| `POST` | `/paises` | `pais:crear` | Alta país. |
| `PUT` | `/paises/:paisId` | `pais:editar` | Modificación país. |
| `PATCH` | `/paises/:paisId/deshabilitar` | `pais:editar` | Desactivar país. |
| `PATCH` | `/paises/:paisId/habilitar` | `pais:editar` | Activar país. |
| `GET` | `/sucursales` | `sucursal:ver` | Listado sucursales. |
| `GET` | `/sucursales/:sucursalId` | `sucursal:ver` | Detalle sucursal. |
| `POST` | `/sucursales` | `sucursal:crear` | Alta sucursal. |
| `PUT` | `/sucursales/:sucursalId` | `sucursal:editar` | Edición sucursal. |
| `PATCH` | `/sucursales/:sucursalId/habilitar` | `sucursal:editar` | Activar sucursal. |
| `PATCH` | `/sucursales/:sucursalId/deshabilitar` | `sucursal:editar` | Desactivar sucursal. |

### Gestión de Salas (Infraestructura)
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/salas` | `sala:ver` | Lista todas las salas. |
| `GET` | `/salas/uso` | `sala:ver` | Lista salas con estado de ocupación actual. |
| `GET` | `/salas/:salaId` | `sala:ver` | Detalle de sala. |
| `POST` | `/salas` | `sala:crear` | Crear sala nueva. |
| `PUT` | `/salas/:salaId` | `sala:editar` | Editar configuración de sala. |
| `PATCH` | `/salas/:salaId/habilitar` | `sala:editar` | Habilitar sala. |
| `PATCH` | `/salas/:salaId/deshabilitar` | `sala:editar` | Deshabilitar sala. |
| `DELETE` | `/salas/:salaId/eliminar` | `sala:eliminar` | Eliminar sala. |

### Control de Tiempos (Acciones)
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `POST` | `/acciones/salas` | `sala:controlar` | **INICIO**. Asigna tiempo a una sala -> Dispara RabbitMQ Start. |
| `PATCH` | `/acciones/salas/pausar/:salaId` | `sala:controlar` | **PAUSA**. Detiene cronómetro -> Dispara RabbitMQ Pause. |
| `PATCH` | `/acciones/salas/reanudar/:salaId` | `sala:controlar` | **PLAY**. Reinicia cronómetro -> Dispara RabbitMQ Resume. |
| `PATCH` | `/acciones/salas/incrementar/:salaId` | `sala:controlar` | Agrega tiempo extra a una sesión activa. |
| `PATCH` | `/acciones/salas/cancelar/:salaId` | `sala:controlar` | Cancela sesión actual. |

### Logística: Proveedores, Productos y Categorías
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/proveedores` | `proveedor:ver` | Directorio proveedores. |
| `POST` | `/proveedores` | `proveedor:crear` | Alta proveedor. |
| `PUT` | `/proveedores/:id` | `proveedor:editar` | Edición proveedor. |
| `GET` | `/productos` | `producto:ver` | Catálogo global. |
| `GET` | `/productos/stats/topProductos` | `producto:ver` | Estadísticas (Más vendidos). |
| `GET` | `/productos/sucursales` | `producto:ver` | Productos filtrados por stock local. |
| `POST` | `/productos` | `producto:crear` | Alta producto. |
| `PUT` | `/productos/:id` | `producto:editar` | Edición producto. |
| `GET` | `/productos-categorias` | `categoria:ver` | Lista categorías. |
| `POST` | `/productos-categorias` | `categoria:crear` | Alta categoría. |

### Inventario: Compras y Stocks
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/compras` | `compra:ver` | Historial órdenes de compra. |
| `POST` | `/compras` | `compra:crear` | Nueva orden de compra. |
| `POST` | `/compras/:id/completar` | `compra:procesar` | Recepción de mercadería (Entrada Stock). |
| `GET` | `/inventario` | `inventario:ver` | Consulta de existencias. |
| `GET` | `/inventario/ajustes` | `ajuste_inventario:ver` | Historial ajustes manuales. |
| `POST` | `/inventario/ajustes` | `ajuste_inventario:crear` | Realizar ajuste manual (+/-). |
| `GET` | `/inventario/transferencias` | `transferencia:ver` | Historial movimientos entre almacenes. |
| `POST` | `/inventario/transferencias` | `transferencia:crear` | Mover stock. |
| `GET` | `/ubicaciones` | `ubicacion:ver` | Lista ubicaciones físicas (Alnacén, Vitrina). |

### Ventas y Caja
| Método | Endpoint | Permiso Requerido | Descripción |
| :--- | :--- | :--- | :--- |
| `GET` | `/ventas` | `venta:ver` | Historial de tickets. |
| `POST` | `/ventas` | `venta:crear` | Generar nueva venta (Checkout). |
| `POST` | `/ventas/:id/pagar` | `venta:cobrar` | Registrar pago parcial/total. |
| `POST` | `/ventas/:id/anular` | `venta:anular` | Revertir venta y devolver stock. |
| `GET` | `/metodos-pago` | `metodo_pago:ver` | Lista formas de pago (Efectivo, QR). |
| `GET` | `/reportes/ventas` | `venta:ver` | PDF Resumen periodo. |
| `GET` | `/ventas/:id/comprobante` | `venta:ver` | PDF Ticket individual. |

### API Usuario Sucursal (`/usuario-sucursal`)
Sub-API simplificada para terminales POS o empleados con permisos limitados.
- `GET /salas`: Vista rápida de salas.
- `POST /acciones/salas`: Control rápido de tiempo.
- `PATCH /acciones/...`: Comandos de pausa/play/cancelar.

### WebSockets (`/ws/v1`)
| Endpoint | Descripción |
| :--- | :--- |
| `/salas` | Monitoreo en tiempo real de todas las salas (Dashboard Admin). |
| `/salas/:salaId` | Monitoreo específico de una sala. |
| `/usuario-sucursal/salas` | Monitoreo para vista POS. |

## 4. Flujo Clave: Venta de Tiempo
1. El operador selecciona una sala y un cliente.
2. Llama a `POST /api/v1/acciones/salas`.
3. El servicio crea un registro en `uso_sala`.
4. El servicio publica un evento en RabbitMQ para que el hardware se encienda.
5. Al finalizar (o cancelar), se calcula el `costo_tiempo` y se genera una `venta` pendiente de pago.

## 5. Base de Datos (Tablas Clave)
- Organizacion: `pais`, `sucursal`.
- Salas: `sala`, `uso_sala`.
- Productos: `producto`, `categoria_producto`, `producto_sucursal`, `ubicacion`.
- Operaciones: `compra`, `inventario`, `transferencia`, `ajuste_inventario`.
- Finanzas: `venta`, `detalle_venta`, `venta_pago`, `metodo_pago`.
