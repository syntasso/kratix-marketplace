import kratix_sdk as ks
import yaml

def create_db(name, replicas, version):
    """
    create a request for a MongoDB database
    """
    mongo = {
        "apiVersion": "mongodbcommunity.mongodb.com/v1",
        "kind": "MongoDBCommunity",
        "metadata": {
            "name": name,
            "namespace": "default"
        },
        "spec": {
            "members": replicas,
            "type": "ReplicaSet",
            "version": version,
            "security": {
                "authentication": {
                    "modes": ["SCRAM"]
                }
            },
            "users": [
                {
                    "name": "my-user",
                    "db": "admin",
                    "passwordSecretRef": {
                        "name": f"mongodb-password-{name}"
                    },
                    "roles": [
                        {"name": "root", "db": "admin"},
                        {"name": "dbAdminAnyDatabase", "db": "admin"},
                        {"name": "clusterAdmin", "db": "admin"},
                        {"name": "userAdminAnyDatabase", "db": "admin"},
                    ],
                    "scramCredentialsSecretName": "my-scram"
                }
            ],
            "additionalMongodConfig": {
                "storage.wiredTiger.engineConfig.journalCompressor": "zlib"
            }
        }
    }
    return mongo

def create_secret(name, password):
    """
    Create a secret for the database
    """
    secret = {
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {
            "name": name,
            "namespace": "default"
        },
        "type": "Opaque",
        "stringData": {
            "password": password,
        }
    }

    return secret

def main():
    # Initialise the sdk
    sdk = ks.KratixSDK()

    # Read the resource input
    resource = sdk.read_resource_input()

    # Get the name the resource
    name = resource.get_name()

    print("Creating mongodb with name:", name)
    
    # Set default values
    replicas= 1
    password= "supersecret"

    # Get correct mongodb version
    major_version = resource.get_value("spec.majorVersion", default=6)

    if major_version == 4:
        version = "4.4.23"
    elif major_version == 5:
        version = "5.0.14"
    elif major_version == 6:
        version = "6.0.10"
    else:
        print(f"Unsupported major version {major_version}, defaulting to 6")
        version = "6.0.10"

    version = resource.get_value("spec.version", default=major_version)
    print("Creating mongodb with version:", version)

    # Write mongodb request and secret to output
    data = yaml.safe_dump(create_db(name, replicas, version)).encode("utf-8")
    sdk.write_output("mongodb-instance.yaml", data)
    data = yaml.safe_dump(create_secret(f"mongodb-password-{name}", password)).encode("utf-8")
    sdk.write_output("secret.yaml", data)

    status = ks.Status()
    status.set("message", f"mongodb version: {version}")
    status.set("version", version)
    status.set("adminUserCredentials.name", f"mongodb-password-{name}")
    status.set("adminUserCredentials.namespace", "default")
    sdk.write_status(status)

if __name__ == "__main__":
    main()
