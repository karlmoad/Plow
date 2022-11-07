# Plow - Snowflake Target 
## Validation

---

### Object Existence Validation
The Snowflake target is programed to validate the existence of an object definition within the target instance to 
enable the application pipeline to determine the correct modification(s) to apply. By setting the ***checkExists*** 
flag on the object definition enrolls the object definition into the validation runtime.  

```yaml
definitionStyle: snowflake
type: table
object:
  name: EXAMPLE_TABLE
  database: DEMO
  schema: PUBLIC
options:
  checkExists: True
spec:
 ...
```

#### Object Types Currently Supported:
- ***Databases***
- ***Schemas***
- ***Tables***
- ***Views***
---

### Structural Change Validation
For specific object types modifications can be applied in place of drop and replace mechanisms by defining the 
object's specification [spec] to have a modification component, see object type specifications for details.  The 
Structural Change Validator will attempt to validate the modification against the current object definition within the 
target and the specification by creating clone objects within validation schemas within tools operating schema.  
Once the clone objects are created and changes applied a structure compare in performed using the Snowflake 
INFORMATION_SCHEMA information to determine alignment.  By setting the ***validation*** flag on the object 
definition it will enroll the object into this process if the type is applicable.

```yaml
definitionStyle: snowflake
type: table
object:
  name: EXAMPLE_TABLE
  database: DEMO
  schema: PUBLIC
options:
  checkExists: True
  validate: True
spec:
 ...
```

#### Object Types Currently Supported:
- ***Tables***
---









