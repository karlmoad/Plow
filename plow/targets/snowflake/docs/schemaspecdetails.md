# Plow - Snowflake Target

## Schema Object Definition Specification

```yaml
definitionStyle: snowflake
type: schema
object:
  name: PUBLIC
  database: DEMO
options:
  checkExists: True
spec:
  owner:
    type: role
    id: DEMO_OWNER
  usage:
    grants:
      - type: role
        id: MIDDLE_EARTH_READ

```