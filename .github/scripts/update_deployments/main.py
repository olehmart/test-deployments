import os
import yaml
import sys
from git import Repo


#################
# ENV Variables #
#################
CONFIG_FILES_LIST = os.environ.get("CONFIG_FILES_LIST").split(" ")
PREV_ENV = os.environ.get("PREV_ENV")
NEXT_ENV = os.environ.get("NEXT_ENV")
#############
# Variables #
#############
destination_branch_name = "auto-deployment-{}".format(NEXT_ENV)
configs_updated = False


#######################
# ENV variables check #
#######################
required_env_variables = ["CONFIG_FILES_LIST", "NEXT_ENV"]
for variable in required_env_variables:
    if variable not in os.environ or os.environ.get(variable) == "":
        print("ERROR! {} variable is not set. Exiting...".format(variable))
        sys.exit(1)

############################
# Get deployments repo #
############################
deployments_repo = Repo("./")
deployments_repo.git.config("--global", "user.email", "noreply@github.com")
deployments_repo.git.config("--global", "user.name", "GitHub BOT")


#########################################
# Checking out to a new/existing branch #
#########################################
branch_exist = False
for ref in deployments_repo.references:
    if "origin/{}".format(destination_branch_name) == ref.name:
        branch_exist = True
if not branch_exist:
    print("Creating a new branch: {} ...".format(destination_branch_name))
    deployments_repo.git.checkout('-b', destination_branch_name)
else:
    print("Branch {} is already exist ...".format(destination_branch_name))
    deployments_repo.git.checkout(destination_branch_name)

#########################
#   Updating versions   #
#########################
for config in CONFIG_FILES_LIST:
    config_file_name = config.split("/")[1]
    if not os.path.exists("{}/{}".format(NEXT_ENV, config_file_name)):
        print("[{0}/{1}]: Skipping update, config file for {0} env doesn't exist...".format(NEXT_ENV, config_file_name))
        continue
    with open(config) as f:
        prev_env_config = yaml.load(f, Loader=yaml.FullLoader)
    with open("{}/{}".format(NEXT_ENV, config_file_name)) as f:
        next_env_config = yaml.load(f, Loader=yaml.FullLoader)
    if prev_env_config["artifact_version"] == next_env_config["artifact_version"]:
        print("[{0}/{1}]: Skipping update, artifact versions are the same in both environments...".format(
            NEXT_ENV, config_file_name))
    else:
        print("[{0}/{1}]: Updating version to {2}...".format(
            NEXT_ENV, config_file_name, prev_env_config["artifact_version"]))
        next_env_config["artifact_version"] = prev_env_config["artifact_version"]
        with open("{}/{}".format(NEXT_ENV, config_file_name), 'w') as f:
            data = yaml.dump(next_env_config, f)
        ###################
        # Staging changes #
        ###################
        deployments_repo.git.add('{}/{}'.format(NEXT_ENV, config_file_name))
        configs_updated = True

#####################
# Creating a commit #
#####################
if configs_updated:
    deployments_repo.git.commit("-m", "auto: propagating {} changes to {}".format(PREV_ENV, NEXT_ENV))

#######################
# Pushing the changes #
#######################
if configs_updated:
    print("Pushing changes ...")
    deployments_repo.git.push("--set-upstream", "origin", destination_branch_name)

os.system('echo "::set-output name=configs_updated::{}"'.format(configs_updated))
