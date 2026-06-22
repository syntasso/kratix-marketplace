# namespace-claim internals

Build the pipeline image:

    cd namespace-claim
    PIPELINE_DIR=configure-pipeline ./internal/scripts/pipeline-image build load

Run unit tests (no cluster or SDK install needed):

    python3 ./internal/configure-pipeline/test/test_claim.py
    ./internal/configure-pipeline/test/test-promise.sh

Run the lifecycle integration test (single-cluster; needs a Kratix platform
with the `namespace` promise installed):

    ./internal/scripts/test
