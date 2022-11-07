# Plow
Change integration tool for database systems. Integrates with git repositories to manage lifecycle of changes to the 
object definitions within the target database system.  The tool utilizes YAML definitions to manage each object's 
definition within the code repository and application to the target. 

Object specifications differ between target system types depending on programming and capabilities, see 
Target System documentation for details. 

This tool requires aspects such as a user account and tracking schema to be established prior to execution.  This 
too is dependent on the target platform.  See target documentation for details

### Logging and Tracking
This tool will attempt to store tracking information for every object state point it endeavors to apply to the 
target.  This information along with a commit tracking information is provided for informational purposes for issue 
followup, but also utilized in operations to determine level of commit alignment with the code repository the 
target exists.  


### Commit Tracking Vs. Fast Forward
***By default***, this system will track the commits applied to the git branch identified in configuration.  In tandem 
with 
the tracking log from the configured target the system determines what commits have not been applied and builds 
change sets (bundles) for each.  Each change set includes only the files modified in that commit.

Alternatively this system can be run with the ***Fast Forward*** option that will advance the to a specific commit
(either the head of the branch or identified by the user) and process all files at that point in the 
branch regardless of commit history. 

```shell
$ plow apply --fast-forward
```

Processing up to and including a specific commit can be achieved by providing a commit id to the ***--commit*** flag 
option 
on applicable commands.  this works with fast-forwarding as well as default commit tracking. ***when not supplied the 
tool assumes the current HEAD of the configured branch to be the designated commit.***

```shell
$ plow list changes --commit=<commit id>
```

To see the commit history of the configured branch execute the following command
```shell
$ plow list commits
```

git 
### Validation
The tool is capable of validating changes to be applied to the target prior to application.  Object specifications 
can take advantage of validation to apply modifications to an existing object in place of drop and replace actions, 
if the target supports the feature.  Additional validation methods can be added to the pipeline by the target's 
programming to provide object state information to the operation.  

Validation will be provided in the command outputs where applicable. If any critical issues are identified in 
the validation stage, the code item is omitted from the change set applied to the target.  It is advised that the 
validation command be run prior to attempting to apply changes. 

```shell
$ plow validate
```

### Currently Supported Target Systems:
| Target System | Validation Details                                 | Object Specification Details                          | Setup                                         |
|:--------------|:---------------------------------------------------|:------------------------------------------------------|:----------------------------------------------|
| **Snowflake** | [link](/plow/targets/snowflake/docs/validation.md) | [link](/plow/targets/snowflake/docs/specification.md) | [link](/plow/targets/snowflake/docs/setup.md) |
