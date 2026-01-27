# Documentación del Servicio: Auth Service

El `Auth Service` es el encargado de la autenticación y autorización dentro del ecosistema Multiroom. Utiliza un modelo de Control de Acceso Basado en Roles (RBAC) para proteger los recursos del sistema.

## 1. Responsabilidades
- Gestión de identidades (Usuarios finales, Administradores, Personal de sucursal).
- Autenticación multifactor (proyectada) y login tradicional.
- Gestión del ciclo de vida de Roles y Permisos.
- Generación y validación de tokens JWT.
- Control de acceso mediante middleware de permisos.

## 2. Tecnologías Específicas
- **JWT**: Firma de tokens con HS256.
- **Bcrypt**: Hashing de contraseñas.
- **Rate Limiter**: Protección contra ataques de fuerza bruta en endpoints de login.

## 3. Listado Completo de Endpoints (`/api/v1`)

### Autenticación e Identidad

| Método | Endpoint               | Permiso Requerido | Descripción                                       |
| :----- | :--------------------- | :---------------- | :------------------------------------------------ |
| `POST` | `/auth/login`          | Público           | Login para usuarios finales (clientes).           |
| `POST` | `/auth/admin/login`    | Público           | Login para administradores globales.              |
| `POST` | `/auth/sucursal/login` | Público           | Login para personal de sucursal.                  |
| `GET`  | `/auth/verify`         | Token Válido      | Verifica la validez del token de usuario cliente. |
| `GET`  | `/auth/admin/verify`   | Token Admin       | Verifica la validez del token de administrador.   |

### Gestión de Roles (`/roles`)

| Método | Endpoint        | Permiso Requerido | Descripción                             |
| :----- | :-------------- | :---------------- | :-------------------------------------- |
| `GET`  | `/roles`        | `rol:ver`         | Lista todos los roles disponibles.      |
| `GET`  | `/roles/:rolId` | `rol:ver`         | Obtiene detalles de un rol específico.  |
| `POST` | `/roles`        | `rol:crear`       | Crea un nuevo rol.                      |
| `PUT`  | `/roles/:rolId` | `rol:editar`      | Modifica un rol existente.              |

### Gestión de Permisos (`/permisos`)

> **Nota**: Endpoints de meta-seguridad.

| Método | Endpoint                | Permiso Requerido   | Descripción                                      |
| :----- | :---------------------- | :------------------ | :----------------------------------------------- |
| `GET`  | `/permisos`             | `permiso:ver`       | Lista todos los permisos del sistema.            |
| `GET`  | `/permisos/:permisoId`  | `permiso:ver`       | Detalle de un permiso.                           |
| `POST` | `/permisos`             | `permiso:gestionar` | Crea/Registra un nuevo permiso en el sistema.    |
| `PUT`  | `/permisos/:permisoId`  | `permiso:gestionar` | Modifica la definición de un permiso.            |

### Gestión de Usuarios Administradores (`/usuariosAdmin`)

| Método | Endpoint                    | Permiso Requerido      | Descripción                         |
| :----- | :-------------------------- | :--------------------- | :---------------------------------- |
| `GET`  | `/usuariosAdmin`            | `usuario_admin:ver`    | Lista administradores del sistema.  |
| `GET`  | `/usuariosAdmin/:usuarioId` | `usuario_admin:ver`    | Detalle de un administrador.        |
| `POST` | `/usuariosAdmin`            | `usuario_admin:crear`  | Registra un nuevo administrador.    |
| `PUT`  | `/usuariosAdmin/:usuarioId` | `usuario_admin:editar` | Edita datos de un administrador.    |

### Gestión de Usuarios Finales (`/usuarios`)

| Método  | Endpoint                           | Permiso Requerido | Descripción                                    |
| :------ | :--------------------------------- | :---------------- | :--------------------------------------------- |
| `GET`   | `/usuarios`                        | `usuario:ver`     | Lista usuarios finales (clientes registrados). |
| `GET`   | `/usuarios/:usuarioId`             | `usuario:ver`     | Detalle de usuario.                            |
| `POST`  | `/usuarios`                        | `usuario:crear`   | Registra manualmente un usuario.               |
| `PATCH` | `/usuarios/:usuarioId/deshabilitar`| `usuario:editar`  | Desactiva una cuenta de usuario.               |
| `PATCH` | `/usuarios/:usuarioId/habilitar`   | `usuario:editar`  | Reactiva una cuenta de usuario.                |

### Gestión de Personal de Sucursal (`/usuariosSucursal`)

| Método  | Endpoint                                   | Permiso Requerido         | Descripción                          |
| :------ | :----------------------------------------- | :------------------------ | :----------------------------------- |
| `GET`   | `/usuariosSucursal`                        | `personal_sucursal:ver`   | Lista empleados de sucursal.         |
| `GET`   | `/usuariosSucursal/:usuarioId`             | `personal_sucursal:ver`   | Detalle de empleado.                 |
| `POST`  | `/usuariosSucursal`                        | `personal_sucursal:crear` | Contrata/Registra un nuevo empleado. |
| `PUT`   | `/usuariosSucursal/:usuarioId`             | `personal_sucursal:editar`| Modifica datos de empleado.          |
| `PATCH` | `/usuariosSucursal/:usuarioId/deshabilitar`| `personal_sucursal:editar`| Suspende acceso a empleado.          |
| `PATCH` | `/usuariosSucursal/:usuarioId/habilitar`   | `personal_sucursal:editar`| Restaura acceso a empleado.          |

## 4. Middleware de Seguridad
El servicio exporta middlewares críticos que son utilizados por otros microservicios (o replicados en lógica) para validar la identidad:
- `VerifyUsuarioAdmin`: Asegura que el token pertenezca a un administrador.
- `VerifyPermission(permiso)`: Valida que el usuario autenticado posea el permiso requerido en su rol para ejecutar la acción.

## 5. Base de Datos
Interactúa principalmente con las tablas:
- `usuario`, `usuario_admin`, `usuario_sucursal`.
- `rol`, `permiso`, `rol_permiso`, `usuario_admin_rol`.
