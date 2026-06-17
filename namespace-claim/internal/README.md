# namespace-claim internals

Build the pipeline image:

    cd namespace-claim
    PIPELINE_DIR=configure-pipeline ./internal/scripts/pipeline-image build load

Run unit tests (no cluster needed):

    ./internal/configure-pipeline/test/test-claim.sh
    ./internal/configure-pipeline/test/test-promise.sh

Run the lifecycle integration test (single-cluster; needs a Kratix platform
with the `Namespace` promise installed):

    ./internal/scripts/test
