# Plow - Snowflake Target

## Database Object Definition Specification

```yaml
definitionStyle: snowflake
type: database
object:
  name: DEMO
options:
  checkExists: true
spec:
  owner:
    type: role
    id: DEMO_OWNER
  usage:
    grants:
      - type: role
        id: DEMO_READ

```