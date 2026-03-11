# PRD - Java Environment Manager (jem)

## 1. Resumen Ejecutivo

**Nombre del proyecto:** `java-env-manager`  
**Alias corto:** `jem`

**Propósito:** Herramienta CLI interactiva para gestionar versiones de JDK y Gradle en entornos de desarrollo local, permitiendo switch rápido entre versiones y descarga automática de las mismas.

**Problema que resuelve:**
- Diferentes proyectos requieren diferentes versiones de JDK/Gradle
- Cambiar manualmente variables de entorno es tedioso y propenso a errores
- No existe una herramienta unificada para gestionar ambos (JDK + Gradle)
- Dificultad para probar código en múltiples versiones
- Los cambios de versión no persisten entre sesiones de shell
- JDKs y Gradles ya instalados no son detectados automáticamente

---

## 2. Objetivos

| Objetivo | Métrica de éxito |
|----------|------------------|
| Gestionar versiones de JDK | Poder instalar, listar, y cambiar entre versiones |
| Gestionar versiones de Gradle | Ídem para Gradle |
| Soporte multi-OS | Funciona en Windows y Linux |
| UX fluida | CLI interactiva con prompts claros |
| Switch rápido | Cambio de versión en < 3 comandos |

---

## 3. Stakeholders

- **Desarrolladores Java** (principal)
- **Equipos de DevOps** (configuración de entornos)

---

## 4. Requisitos Funcionales

### 4.1 Gestión de JDK

| ID | Requisito | Prioridad |
|----|-----------|-----------|
| JDK-01 | Listar versiones de JDK instaladas (gestionadas por jem) | Alta |
| JDK-02 | Listar versiones de JDK disponibles para descargar | Media |
| JDK-03 | Descargar e instalar una versión de JDK | Alta |
| JDK-04 | Cambiar (switch) la versión activa de JDK | Alta |
| JDK-05 | Mostrar versión activa actual | Alta |
| JDK-06 | Detectar automáticamente JDKs existentes en el sistema | Alta |
| JDK-07 | Registrar JDKs detectados para su gestión | Alta |
| JDK-08 | Soportar múltiples proveedores de JDK (Temurin, Zulu, Corretto, etc.) | Alta |

### 4.2 Gestión de Gradle

| ID | Requisito | Prioridad |
|----|-----------|-----------|
| GRAD-01 | Listar versiones de Gradle instaladas (gestionadas por jem) | Alta |
| GRAD-02 | Listar versiones de Gradle disponibles para descargar | Media |
| GRAD-03 | Descargar e instalar una versión de Gradle | Alta |
| GRAD-04 | Cambiar (switch) la versión activa de Gradle | Alta |
| GRAD-05 | Mostrar versión activa actual | Alta |
| GRAD-06 | Detectar automáticamente Gradles existentes en el sistema | Alta |
| GRAD-07 | Registrar Gradles detectados para su gestión | Alta |

### 4.3 Sistema

| ID | Requisito | Prioridad |
|----|-----------|-----------|
| SYS-01 | Soporte para Windows | Alta |
| SYS-02 | Soporte para Linux | Alta |
| SYS-03 | Configuración persistente de la versión activa | Alta |
| SYS-04 | Integración con PATH del sistema/shell | Alta |

### 4.4 Persistencia y PATH Management

| ID | Requisito | Prioridad |
|----|-----------|-----------|
| PER-01 | Persistir versión activa de JDK entre sesiones de shell | Alta |
| PER-02 | Persistir versión activa de Gradle entre sesiones de shell | Alta |
| PER-03 | Directorio `~/.jem/current/` con symlinks a versiones activas de JDK/Gradle | Alta |
| PER-04 | Comando `jem init` para configurar environment variables en el shell actual | Alta |
| PER-05 | El cambio de versión actualiza symlinks automáticamente | Alta |
| PER-06 | No requiere scripts de inicialización por shell | Alta |
| PER-07 | Funciona automáticamente en todas las sesiones nuevas | Alta |
| PER-08 | Manejo de JAVA_HOME (symlink o variable de entorno) | Media |
| PER-09 | Shell function wrapper para `jem use` automático (sin eval manual) | Alta |
| PER-10 | Soporte Bash/Zsh (Linux/macOS) y PowerShell (Windows) para wrapper | Alta |

---

## 5. Requisitos No Funcionales

| ID | Requisito |
|----|-----------|
| NFR-01 | Interfaz CLI interactiva (prompts, menús) |
| NFR-02 | Tiempo de respuesta < 2s para operaciones locales |
| NFR-03 | Manejo graceful de errores de red |
| NFR-04 | Logging de operaciones para debugging |
| NFR-05 | Instalación limpia (sin modificar sistema globalmente) |

---

## 6. Restricciones y Suposiciones

**Restricciones:**
- No requiere privilegios de administrador/root para uso básico
- Descargas desde fuentes oficiales (Adoptium/Eclipse Temurin para JDK, Gradle official para Gradle)

**Suposiciones:**
- El usuario tiene conexión a internet para descargas
- El usuario tiene permisos de escritura en el directorio de instalación

---

## 7. Out of Scope (v1)

- macOS support (futuro)
- Maven support (futuro)
- Gestión de JVM args personalizadas
- Integración con IDEs
- Gestión de múltiples shells simultáneos

---

## 8. Arquitectura Preliminar

```
java-env-manager/
├── cmd/                 # CLI entrypoints
├── internal/
│   ├── jdk/            # Lógica de gestión JDK
│   ├── gradle/         # Lógica de gestión Gradle
│   ├── config/         # Configuración persistente
│   ├── platform/       # Abstracción OS (Windows/Linux)
│   └── downloader/     # Lógica de descarga
├── pkg/
│   └── interactive/    # UI interactiva (prompts)
└── go.mod
```

**Lenguaje sugerido:** Go (binario único, fácil distribución, cross-compile)

---

## 9. Comandos CLI Propuestos

```bash
jem                         # Menú interactivo principal
jem init                    # Configura environment variables para el shell actual
jem setup                   # Configura shell para usar jem init automáticamente
jem list jdk                # Lista JDKs instalados
jem list gradle             # Lista Gradles instalados
jem install jdk <version>   # Instala JDK
jem install gradle <version> # Instala Gradle
jem use jdk <version>       # Switch a versión JDK (actualiza symlink)
jem use gradle <version>    # Switch a versión Gradle (actualiza symlink)
jem current                 # Muestra versiones activas
jem scan                    # Detecta JDKs y Gradles existentes en el sistema
```

### 9.1 Estrategia de PATH Management

**Estructura de directorios:**
```
~/.jem/
├── current/                # Symlinks a versiones activas (usado por jem init)
│   ├── java -> ../jdks/temurin-21
│   └── gradle -> ../gradles/8.5
├── jdks/
│   ├── temurin-17/        # JDK 17 Temurin
│   ├── corretto-21/       # JDK 21 Corretto
│   └── current -> temurin-17/   # Symlink a versión activa (legacy)
├── gradles/
│   ├── gradle-8.5/
│   ├── gradle-8.6/
│   └── current -> gradle-8.5/   # Symlink a versión activa (legacy)
└── config.toml             # Estado y configuración
```

**Comportamiento:**
1. `jem init` - Genera exports de JAVA_HOME, GRADLE_HOME y PATH basados en `~/.jem/current/`
2. `jem use jdk <version>` - Actualiza el symlink `~/.jem/current/java`
3. `jem use gradle <version>` - Actualiza el symlink `~/.jem/current/gradle`
4. El shell evalúa `jem init` al iniciar para configurar el environment

**Ventajas:**
- Soporte para session-only switches (solo en la sesión actual)
- Soporte para default versions (persistente entre sesiones)
- JAVA_HOME y GRADLE_HOME configurados automáticamente
- Funciona en bash, zsh, y PowerShell

### 9.2 Comportamiento de `jem init`

**Variables de entorno generadas:**

| Variable | Valor |
|----------|-------|
| JAVA_HOME | `~/.jem/current/java` (symlink al JDK activo) |
| GRADLE_HOME | `~/.jem/current/gradle` (symlink al Gradle activo) |
| PATH | `$JAVA_HOME/bin:$GRADLE_HOME/bin:$PATH` |

### 9.3 Shell Function Wrapper (Automático `jem use`)

**Problema:** `jem` es un proceso binario que no puede modificar las variables de entorno del shell padre. Los usuarios debían ejecutar manualmente `source ~/.zshrc` o `eval "$(jem use jdk 21 --output-env)"`.

**Solución:** Shell function wrapper instalado por `jem setup` que intercepta comandos `jem use` y auto-evalúa el output de `--output-env`.

#### Arquitectura del wrapper

```bash
# Bash/Zsh (~/.bashrc o ~/.zshrc)
jem() {
    case "$1" in
        use)
            shift
            eval "$(command jem use "$@" --output-env)"
            ;;
        *)
            command jem "$@"
            ;;
    esac
}
```

```powershell
# PowerShell ($PROFILE)
function jem {
    param([Parameter(ValueFromRemainingArguments)]$Args)
    if ($Args[0] -eq 'use') {
        $output = & jem $Args --output-env 2>&1
        if ($LASTEXITCODE -eq 0) {
            Invoke-Expression $output
        } else {
            Write-Host $output
        }
    } else {
        & jem @Args
    }
}
```

#### Flujo de ejecución

```
Usuario: jem use jdk 21
    │
    └─► Shell wrapper intercepta
            │
            ├─► Ejecuta: command jem use jdk 21 --output-env
            │
            ├─► jem binario devuelve:
            │       export JAVA_HOME="~/.jem/jdks/21-amzn"
            │       export PATH="~/.jem/jdks/21-amzn/bin:$PATH"
            │
            └─► Wrapper evalúa output → entorno actualizado
```

#### Soporte por Shell

| Shell | Soporte | Notas |
|-------|---------|-------|
| Bash | ✅ | Linux |
| Zsh | ✅ | Linux/macOS |
| PowerShell | ✅ | Windows |
| Fish | ❌ | Limitación del shell - usar `jem use default` |

#### Estado de Validación

| Plataforma | Estado |
|------------|--------|
| Linux (Bash) | ✅ Validado manualmente |
| Linux (Zsh) | ✅ Validado manualmente |
| Windows (PowerShell) | ⏳ Pendiente de validación |
| Error handling | ✅ Validado - entorno intacto en errores |

#### Migración para usuarios existentes

1. Ejecutar `jem setup` nuevamente
2. `source ~/.zshrc` o reiniciar terminal
3. `jem use jdk 21` funciona inmediatamente

**Archivos de configuración por shell:**

| OS | Shell | Archivo |
|----|-------|---------|
| Linux | bash | `~/.bashrc` |
| Linux | zsh | `~/.zshrc` |
| Windows | PowerShell | `$PROFILE` |
| Windows | cmd | Registro de usuario (no recomendado, mejor PowerShell) |

---

## 10. Fuentes de Descarga

### 10.1 Proveedores de JDK

| Proveedor | Distribución | URL | Notas |
|-----------|-------------|-----|-------|
| Eclipse Temurin | Adoptium | https://adoptium.net/ | Default, OpenJDK |
| Amazon Corretto | Amazon | https://aws.amazon.com/corretto/ | OpenJDK, optimizado AWS |
| Azul Zulu | Azul Systems | https://www.azul.com/downloads/ | OpenJDK, múltiple variantes |
| Microsoft Build of OpenJDK | Microsoft | https://learn.microsoft.com/java/openjdk/ | OpenJDK para Azure |
| Oracle GraalVM | Oracle | https://www.graalvm.org/ | JDK con AOT compilation |
| BellSoft Liberica | BellSoft | https://bell-sw.com/ | OpenJDK, incluye JavaFX |
| IBM Semeru | IBM | https://developer.ibm.com/languages/java/semeru-runtimes/ | OpenJDK con OpenJ9 |
| IBM SDK (WebSphere) | IBM | https://www.ibm.com/docs/en/sdk-java-technology | JDK para WebSphere Application Server |

### 10.2 Gradle

| Fuente | URL |
|--------|-----|
| Gradle Official | https://gradle.org/releases/ |

---

## 11. Roadmap Preliminar

| Fase | Entregable |
|------|------------|
| Fase 1 | Core: Gestión de JDK (scan, list, install, use) + Symlinks |
| Fase 2 | Gestión de Gradle (scan, list, install, use) |
| Fase 3 | Múltiples proveedores de JDK |
| Fase 4 | Menú interactivo completo |
| Fase 5 | Mejoras de UX y estabilidad |

---

## 12. Decisiones Pendientes

### Configuración
- [x] Ubicación por defecto: `~/.jem/`
- [x] Formato de configuración: TOML (config.toml)
- [x] Estrategia de persistencia: Symlinks en `~/.jem/current/` con `jem init` para configurar environment

### PATH Management
- [x] Estrategia de JAVA_HOME: Si no existe → crear, si existe → preguntar para reemplazar
- [x] Estrategia de GRADLE_HOME: Ídem que JAVA_HOME
- [x] Estrategia de PATH: `jem init` genera exports para JAVA_HOME/bin y GRADLE_HOME/bin
- [x] jem tiene prioridad sobre instalaciones locales
- [x] Archivos por OS/shell: `.bashrc`, `.zshrc`, `$PROFILE` (ver sección 9.2)

### Proveedores
- [x] Proveedor por defecto: Eclipse Temurin
- [x] Formato: `jem install jdk <version> --provider <nombre>`
- [x] Versiones: Especificar versión exacta, `--lts` para última LTS, `--latest` para última disponible
- [x] Builds: Soportar versión específica (21.0.2) o major version (21)

### Detección
- [x] Rutas de detección automática por OS (ver sección 12.1)
- [x] Distinción gestionados vs externos: Ubicación + symlink para importar
- [x] Importación: `jem import jdk <ruta> --name <nombre>` (crea symlink)

---

## 12.1 Rutas de Detección Automática

### JDK - Windows

| Proveedor | Ruta |
|-----------|------|
| Oracle JDK | `C:\Program Files\Java\jdk-*` |
| Eclipse Temurin | `C:\Program Files\Eclipse Adoptium\` |
| Amazon Corretto | `C:\Program Files\Amazon Corretto\` |
| Azul Zulu | `C:\Program Files\Zulu\` |
| Microsoft OpenJDK | `C:\Program Files\Microsoft\jdk-*` |
| IBM SDK | `C:\Program Files\IBM\Java\*` |
| JetBrains/IntelliJ | `%USERPROFILE%\.jdks\` |
| SDKMAN | `%USERPROFILE%\.sdkman\candidates\java\` |

### JDK - Linux

| Proveedor | Ruta |
|-----------|------|
| General (distro) | `/usr/lib/jvm/` |
| Oracle JDK | `/usr/java/` |
| IBM SDK | `/opt/ibm/java-*` |
| JetBrains/IntelliJ | `~/.jdks/` |
| SDKMAN | `~/.sdkman/candidates/java/` |
| Manual | `/opt/java/`, `/usr/local/java/` |

### Gradle - Windows

| Origen | Ruta |
|--------|------|
| Instalación manual | `C:\Gradle\` |
| SDKMAN | `%USERPROFILE%\.sdkman\candidates\gradle\` |
| Gradle Wrapper | `%USERPROFILE%\.gradle\wrapper\dists\` |

### Gradle - Linux

| Origen | Ruta |
|--------|------|
| Instalación manual | `/opt/gradle/`, `/usr/local/gradle/` |
| SDKMAN | `~/.sdkman/candidates/gradle/` |
| Gradle Wrapper | `~/.gradle/wrapper/dists/` |

### Gestión de JDKs externos

**JDKs gestionados por jem:**
- Ubicados en `~/.jem/jdks/`
- Instalados vía `jem install` o importados

**JDKs externos:**
- Detectados en rutas estándar del sistema
- Pueden importarse con: `jem import jdk <ruta> --name <nombre>`
- Importación crea symlink en `~/.jem/jdks/` sin copiar archivos

---

## 13. Consideraciones Técnicas Adicionales

### 13.1 Symlinks en Windows

| Aspecto | Detalle |
|---------|---------|
| **Requisito** | Windows requiere Developer Mode o permisos de administrador para crear symlinks |
| **Fallback** | Usar Junctions (directory junctions) que no requieren privilegios especiales |
| **Detección** | Verificar al inicio si symlinks están disponibles, sino usar junctions |
| **Documentación** | Incluir en `jem setup` instrucciones para habilitar Developer Mode |

### 13.2 Verificación de Integridad de Downloads

| Aspecto | Detalle |
|---------|---------|
| **Riesgo** | Downloads corruptos o modificados pueden causar errores silenciosos |
| **Solución** | Validar checksums SHA256 de cada descarga |
| **Fuentes** | Adoptium publica checksums en su API; Gradle incluye checksums en releases |
| **Comportamiento** | Si checksum falla → reintentar descarga o alertar al usuario |

### 13.3 APIs de Proveedores

| Proveedor | API | Notas |
|-----------|-----|-------|
| Eclipse Temurin | `https://api.adoptium.net/v3/assets/latest/{version}/{architecture}` | JSON con URLs de descarga y checksums |
| Amazon Corretto | Scraping o manifest JSON | Menos documentada |
| Gradle | `https://services.gradle.org/versions/current` y `/all` | JSON con versiones disponibles |

**Estrategia:** Diseñar interfaz `Provider` abstracta para encapsular lógica de cada proveedor.

### 13.4 Soporte para Proxies Corporativos

| Aspecto | Detalle |
|---------|---------|
| **Variables soportadas** | `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` |
| **Autenticación** | Soportar proxies con autenticación básica |
| **Timeouts** | Configurables para entornos con conexión lenta |
| **Mensajes** | Errores claros cuando proxy bloquea conexión |

### 13.5 Librerías Go Recomendadas

| Propósito | Librería | Alternativas |
|-----------|----------|--------------|
| CLI framework | `github.com/spf13/cobra` | `urfave/cli` |
| Config TOML | `github.com/BurntSushi/toml` | `pelletier/go-toml` |
| HTTP client | `net/http` (stdlib) + retry con `github.com/hashicorp/go-retryablehttp` | — |
| Progress bars | `github.com/schollz/progressbar/v3` | `briandowns/spinner` |
| Prompts interactivos | `github.com/AlecAivazis/survey/v2` | `charmbracelet/bubbletea` |
| Logging | `github.com/rs/zerolog` | `slog` (Go 1.21+) |
| Checksums | `crypto/sha256` (stdlib) | — |

### 13.6 Manejo de Errores

| Tipo | Ejemplo | Estrategia |
|------|---------|------------|
| Red | Timeout, conexión rechazada | Retry con backoff, mensaje claro |
| Disco | Espacio insuficiente, permisos | Verificar antes de descargar |
| Validación | Checksum incorrecto, archivo corrupto | Reintentar o abortar con guía |
| Configuración | Archivo config.toml corrupto | Backup automático, recrear defaults |

---

## 14. Dependencias entre Requisitos

```
                    ┌─────────────────────────────────────┐
                    │         SYS-01/02: Soporte OS       │
                    └─────────────────┬───────────────────┘
                                      │
              ┌───────────────────────┼───────────────────────┐
              ▼                       ▼                       ▼
    ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
    │  JDK-06/07      │     │  PER-03/04/05   │     │  GRAD-06/07     │
    │  Detectar JDKs  │     │  Symlinks/PATH  │     │  Detectar Gradle│
    └────────┬────────┘     └────────┬────────┘     └────────┬────────┘
             │                       │                       │
             │               ┌───────┴───────┐               │
             │               ▼               ▼               │
             │     ┌──────────────┐  ┌──────────────┐        │
             │     │  JDK-01/04   │  │ GRAD-01/04   │        │
             │     │  List/Switch │  │ List/Switch  │        │
             │     └──────┬───────┘  └──────┬───────┘        │
             │            │                 │                │
             └────────────┴────────┬────────┴────────────────┘
                                   │
                                   ▼
                         ┌─────────────────┐
                         │  JDK-03 / GRAD-03│
                         │  Instalar       │
                         └─────────────────┘
```

**Orden de implementación sugerido:**
1. Core de plataforma (SYS-01/02) + detección (JDK-06, GRAD-06)
2. Sistema de symlinks y PATH (PER-03/04/05)
3. Comandos list/use (JDK-01/04, GRAD-01/04)
4. Comandos install (JDK-03, GRAD-03) con verificación de integridad