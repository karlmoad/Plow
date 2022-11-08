# Plow - Snowflake Target

## Role Object Definition Specification


```yaml 

definitionStyle: snowflake
type: role
spec:
  roles:
    - DEMO_OWNER
    - DEMO_READ
    - DEMO_RW
  grants:
    - role: DEMO_READ
      object:
        type: role
        id: DEMO_RW
        
```