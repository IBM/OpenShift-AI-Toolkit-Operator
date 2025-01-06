import tritonclient.grpc as grpcclient
import tritonclient.utils as utils

# Server URL
TRITON_SERVER_URL = "localhost:8001"

try:
    triton_client = grpcclient.InferenceServerClient(
        url=TRITON_SERVER_URL,
        ssl=False,
    )

    # Check server health
    if triton_client.is_server_live():
        print("Triton server is live.")
    else:
        print("Triton server is not live.")

except utils.InferenceServerException as e:
    print(f"Failed to connect to Triton server: {e}")
