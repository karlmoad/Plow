# Plow - Snowflake Target

## Specifications 

All object yaml specifications contain the same base structure with the ***spec*** section differing by the type of 
object.  The base yaml structure contains the following structure and elements.


```yaml

definitionStyle: snowflake        (required)
type: <object type name>          (required)
object:
  name: <object name>             (required)
  database: <object database> 
  schema: <object schema>
options:
  [Validation options]                 
spec:
  <will vary by object type>


```

| Element         | Comment                                                                                                                                                                                | Required  | Applicable Values                   |
|:----------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------|:------------------------------------|
| definitionStyle | Identifies the target system style the object defninintion conforms                                                                                                                    | Yes  | snowflake                           |
| type            | identifies the object type the defninition details, should be lowercase name                                                                                                           | Yes | [object type list](/objecttypes.md) |
| object.name     | Name identifier of the object within the target                                                                                                                                        | Yes | string  (*)                         |
| object.database | Name of the database in which the object will be defnined, if applicable                                                                                                               | No | string   (*)                        |
| object.schema   | Name of the schema within the database the object will be defnined, if applicable.  Note: If defined object.database becomes a required field or errors will occur                     | No | string   (*)                        |
| options         | This section defines the options present for the object including validation and other pre/post processing, see [option details](/validation.md) for more information                  | No | [option details](/validation.md)    |
| spec | this element will contain the actual definition of the object, the stucture within is dependant on the object type being defnined, see [object type list](/objecttypes.md) for details | Yes | [object type list](/objecttypes.md) |

(*): Alphanumeric value consisting of character set { A-Z , 0-9, _ (underscore) }.  
