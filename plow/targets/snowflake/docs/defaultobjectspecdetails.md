# Plow - Snowflake Target

## Base Object Definition Specification 

Objects defined using the base object specification consists of 4 "scopes" which can executed as part of the 
validation and application phases depending on options configuration within the object definition.  

This object definition style take advantage of using target native SQL commands executed in a defined order.  The 
scopes execute in the order defined below, if an error occurs during application of a scope all following scopes 
will be skipped.  Scopes can include multiple commands to be executed, each command shoudl be terminated with a ";" 
character.  

### Metadata (meta element)

The first element of the base object specification is the meta (metadata) structure. This metadata structure 
identifies the owning role that will be allied to the object within the target.  The owner must be a role, and must be 
defined and present within the system prior to the execution of the objects specification commands.  Execution 
order of object types can be seen [here](/plow/targets/snowflake/docs/objecttypes.md)   

```yaml 
spec:
  meta:
    owner:
      type: role
      id: DEMO_OWNER
```



### Scopes

| Scope / Element Name | Required | Description                                                                                                                                                                                                                                                                                                                                                                                                                                |
|:---------------------| :--- |:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| pre                  | No | If present, command(s) contained within will be applied to the target prior to other scopes within the definition                                                                                                                                                                                                                                                                                                                          |
| init                 | No | If present, command(s) within this scoep will be executed when the object is deemed to not exist via validation and requires instantiation.  If validation is not configured commands within this scope are executed following any ***pre*** scope commands.                                                                                                                                                                               |
| change               | No | Command(s) within this scope are intended to make changes to an existsing object's defninition within the target. For this scope to be applied validation must be configured so the objects state of exisstance can be determined.  If validation is not applied this scope will be ignored, Note: all structural mnodification defnined within this scope shoudl also be refected in the ***init*** scope as well for validation purposes |
| post                 | No | If Present, command(s) within this scope are executed following the application of all other scopes                                                                                                                                                                                                                                                                                                                                        |

Commands can be broken into multiple lines for readability by adding the "|" charater and a line feed following the 
scope element name within the yaml structure.  Please note if the statement contains multiple commands each command 
must be terminated with a ";" character to avoid error.   

### Variables
To provided for the mechanics of the validation system to inject required values to render validation structures, 
variables are used within the scope command statements to reflect values form the header object name, database, 
schema settings.  each variable is defined using {{VARIABLE NAME (NAME,DATABASE, or SCHEMA)}} format using all 
capital letters, in place of the value.  The system will inject the correct value form the header settings to 
generate the final desired effect.    

***If the object definition wants to take advantage of structural validation variable names must be used within the 
scope command definitions***


### Example

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
  meta:
    owner:
      type: role
      id: DEMO_OWNER
  pre: USE DATABASE {{DATABASE}};
  init: |
    CREATE TABLE {{DATABASE}}.{{SCHEMA}}.EXAMPLE_TABLE(
      IDKEY INTEGER,
      NAME VARCHAR(200),
        ....
      AVG_HEIGHT NUMBER(15,2)
    );
  change: |
    ALTER TABLE {{DATABASE}}.{{SCHEMA}}.EXAMPLE_TABLE ADD COLUMN AVG_HEAIGHT NUMBER(15,2);
  post: |
    GRANT SELECT ON TABLE {{DATABASE}}.{{SCHEMA}}.EXAMPLE_TABLE TO ROLE DEMO_READ;

```

