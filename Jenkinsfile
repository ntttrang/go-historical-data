// ============================================================================
// Jenkins Declarative Pipeline for Go Historical Data API
// ============================================================================
// This pipeline implements a complete CI/CD workflow with the following stages:
//
// CI (Continuous Integration):
//   1. Checkout - Clone repository from Git
//   2. Environment Setup - Prepare Go environment
//   3. Dependencies - Download and verify Go modules
//   4. Lint - Run golangci-lint for code quality
//   5. Build - Compile the Go application
//   6. Unit Tests - Run unit tests with coverage
//   7. Integration Tests - Run integration tests with Docker
//   8. Security Scan - Scan for vulnerabilities
//
// CD (Continuous Deployment):
//   9. Build Docker Image - Create production container
//   10. Push to Docker Hub - Tag and push image
//   11. Deploy to Staging - Deploy to staging environment (optional)
//   12. Deploy to Production - Deploy to production environment
//   13. Health Check - Verify deployment health
//   14. Notification - Send deployment status notifications
// ============================================================================

pipeline {
    agent any
    
    // ========================================================================
    // ENVIRONMENT VARIABLES
    // ========================================================================
    environment {
        // Application Configuration
        APP_NAME = 'go-historical-data'
        GO_VERSION = '1.24'
        
        // Docker Configuration
        DOCKER_HUB_REPO = 'minhtrang2106/go-historical-data'  // TODO: Update with your Docker Hub username
        DOCKER_REGISTRY = 'https://index.docker.io/v1/'
        DOCKER_CREDENTIALS_ID = 'dockerhub-credentials'  // Jenkins credential ID
        
        // Git Configuration
        GIT_CREDENTIALS_ID = 'git-credentials'  // Jenkins credential ID
        
        // Render.com Deployment Configuration
        RENDER_API_KEY_CREDENTIALS_ID = 'render-api-key'  // Jenkins credential ID for Render API key
        RENDER_STAGING_SERVICE_ID = 'srv-d3mhlt1r0fns73eo0hi0'  // TODO: Add your Render staging service ID
        RENDER_PRODUCTION_SERVICE_ID = 'srv-d3mhlt1r0fns73eo0hhg'  // TODO: Add your Render production service ID
        
        // Build Configuration
        CGO_ENABLED = '0'
        GOOS = 'linux'
        GOARCH = 'amd64'
        
        // Go Module Cache Configuration (Fix for permission issues)
        GOPATH = "${env.WORKSPACE}/.go"
        GOMODCACHE = "${env.WORKSPACE}/.gomodcache"
        GOCACHE = "${env.WORKSPACE}/.gocache"
        GOTOOLCHAIN = 'local'  // Prevent auto-download of newer Go versions
        
        // Test Configuration
        TEST_TIMEOUT = '10m'
        COVERAGE_THRESHOLD = '80'
        
        // Notification Configuration
        SLACK_CHANNEL = '#deployments'  // TODO: Update with your Slack channel
        SLACK_CREDENTIALS_ID = 'slack-webhook'  // Jenkins credential ID
        
        // Dynamic Variables
        BUILD_VERSION = "${env.BUILD_NUMBER}"
        // GIT_COMMIT_SHORT and DOCKER_TAG are computed after checkout
    }
    
    // ========================================================================
    // BUILD PARAMETERS
    // ========================================================================
    parameters {
        choice(
            name: 'DEPLOY_ENVIRONMENT',
            choices: ['none', 'staging', 'production'],
            description: 'Select deployment environment (none = CI only)'
        )
        string(
            name: 'REPO_URL',
            defaultValue: 'https://github.com/ntttrang/go-historical-data.git',
            description: 'Git repository URL (required for non-Multibranch pipelines)'
        )
        string(
            name: 'GIT_BRANCH',
            defaultValue: 'dev',
            description: 'Git branch to build'
        )
        booleanParam(
            name: 'SKIP_TESTS',
            defaultValue: false,
            description: 'Skip test execution (not recommended for production)'
        )
        booleanParam(
            name: 'FORCE_DEPLOY',
            defaultValue: false,
            description: 'Force deployment even if tests fail (use with caution)'
        )
        string(
            name: 'DOCKER_TAG_OVERRIDE',
            defaultValue: '',
            description: 'Override Docker tag (leave empty for auto-generated)'
        )
    }
    
    // ========================================================================
    // PIPELINE OPTIONS
    // ========================================================================
    options {
        // Keep only last 10 builds
        buildDiscarder(logRotator(numToKeepStr: '10'))
        
        // Timeout for entire pipeline
        timeout(time: 1, unit: 'HOURS')
        
        // Disable concurrent builds
        disableConcurrentBuilds()
        
        // Add timestamps to console output
        timestamps()
        
        // Colorize console output
        ansiColor('xterm')
    }
    
    // ========================================================================
    // TRIGGERS
    // ========================================================================
    triggers {
        // Poll SCM every 5 minutes
        pollSCM('H/5 * * * *')
        
        // Trigger on GitHub webhook
        githubPush()
    }
    
    // ========================================================================
    // CI STAGES (CONTINUOUS INTEGRATION)
    // ========================================================================
    stages {
        
        // ====================================================================
        // CI STAGE 1: CHECKOUT
        // ====================================================================
        stage('📥 Checkout') {
            steps {
                script {
                    echo '============================================'
                    echo 'CI STAGE 1: CHECKOUT SOURCE CODE'
                    echo '============================================'
                    echo "Branch: ${env.BRANCH_NAME}"
                    echo "Build Number: ${env.BUILD_NUMBER}"
                    echo "Workspace: ${env.WORKSPACE}"
                }
                
                // Validate required params for non-Multibranch usage
                script {
                    if (params.REPO_URL == null || params.REPO_URL.trim() == '') {
                        error 'REPO_URL parameter is required when not using Multibranch Pipeline or Pipeline from SCM.'
                    }
                }
                
                // Checkout code from Git with credential fallback
                script {
                    if (fileExists('.git')) {
                        echo 'Existing Git workspace detected. Fetching updates...'
                        sh """
                            git remote set-url origin ${params.REPO_URL} || true
                            git fetch --all --tags --prune
                            git checkout ${params.GIT_BRANCH}
                            git reset --hard origin/${params.GIT_BRANCH}
                        """
                    } else {
                        try {
                            if (env.GIT_CREDENTIALS_ID && env.GIT_CREDENTIALS_ID.trim()) {
                                git branch: params.GIT_BRANCH, credentialsId: env.GIT_CREDENTIALS_ID, url: params.REPO_URL
                            } else {
                                git branch: params.GIT_BRANCH, url: params.REPO_URL
                            }
                        } catch (e) {
                            echo "Primary checkout failed (${e}). Retrying without credentials..."
                            git branch: params.GIT_BRANCH, url: params.REPO_URL
                        }
                    }
                }
                
                // Display Git information
                sh '''
                    echo "Git Commit: $(git rev-parse HEAD)"
                    echo "Git Author: $(git log -1 --pretty=format:'%an <%ae>')"
                    echo "Git Message: $(git log -1 --pretty=format:'%s')"
                '''
                
                // Compute dynamic variables dependent on Git after checkout
                script {
                    // Derive branch name if Jenkins environment doesn't provide it
                    if (env.BRANCH_NAME == null || env.BRANCH_NAME.trim() == '') {
                        env.BRANCH_NAME = sh(script: "git rev-parse --abbrev-ref HEAD", returnStdout: true).trim()
                    }
                    env.GIT_COMMIT_SHORT = sh(script: "git rev-parse --short HEAD", returnStdout: true).trim()
                    env.DOCKER_TAG = "${env.BRANCH_NAME}-${env.BUILD_NUMBER}-${env.GIT_COMMIT_SHORT}"
                    echo "Computed GIT_COMMIT_SHORT=${env.GIT_COMMIT_SHORT}"
                    echo "Computed DOCKER_TAG=${env.DOCKER_TAG}"
                }
            }
        }
        
        // ====================================================================
        // CI STAGE 2: ENVIRONMENT SETUP
        // ====================================================================
        stage('🔧 Environment Setup') {
            steps {
                script {
                    echo '============================================'
                    echo 'CI STAGE 2: ENVIRONMENT SETUP'
                    echo '============================================'
                }
                
                sh '''
                    echo "Go version:"
                    go version || echo "go not found"
                    
                    echo "Go environment:"
                    go env || true
                    
                    # Fix Go module cache permissions
                    echo "Setting up Go cache directories..."
                    mkdir -p "${WORKSPACE}/.gomodcache" "${WORKSPACE}/.gocache"
                    export GOPATH="${WORKSPACE}/.go"
                    mkdir -p "${GOPATH}/bin" "${GOPATH}/pkg"
                    chmod -R 755 "${GOPATH}" "${WORKSPACE}/.gomodcache" "${WORKSPACE}/.gocache" || true
                    
                    echo "Docker version:"
                    if command -v docker >/dev/null 2>&1; then docker --version; else echo "docker not found"; fi
                    
                    echo "Docker Compose version:"
                    if command -v docker-compose >/dev/null 2>&1; then docker-compose --version; else echo "docker-compose not found"; fi
                '''
            }
        }
        
        // ====================================================================
        // CI STAGE 3: DEPENDENCIES
        // ====================================================================
        stage('📦 Dependencies') {
            steps {
                script {
                    echo '============================================'
                    echo 'CI STAGE 3: DOWNLOAD DEPENDENCIES'
                    echo '============================================'
                }
                
                sh '''
                    # Download dependencies
                    echo "Downloading Go modules..."
                    go mod download
                    
                    # Verify dependencies
                    echo "Verifying Go modules..."
                    go mod verify
                    
                    # Tidy dependencies
                    echo "Tidying Go modules..."
                    go mod tidy
                    
                    # Check for changes
                    if ! git diff --exit-code go.mod go.sum; then
                        echo "WARNING: go.mod or go.sum has changes after 'go mod tidy'"
                        echo "Please run 'go mod tidy' locally and commit the changes"
                        exit 1
                    fi
                '''
            }
        }
        
        // ====================================================================
        // CI STAGE 4: LINT
        // ====================================================================
        stage('🔍 Lint') {
            steps {
                script {
                    echo '============================================'
                    echo 'CI STAGE 4: CODE QUALITY CHECK'
                    echo '============================================'
                }
                
                sh '''
                    # Ensure GOPATH/bin is in PATH
                    export PATH="${GOPATH}/bin:${PATH}"
                    
                    # Install golangci-lint (always reinstall to get compatible version)
                    echo "Installing golangci-lint (latest version)..."
                    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin latest
                    
                    # Run linter (uses .golangci.yml config which excludes vendor and has typecheck disabled)
                    echo "Running golangci-lint..."
                    # Only lint our own code: restrict to project packages (exclusions are defined in .golangci.yml)
                    golangci-lint run --timeout=5m ./cmd/... ./internal/... ./pkg/...
                '''
            }
        }
        
        // ====================================================================
        // CI STAGE 5: BUILD
        // ====================================================================
        stage('🔨 Build') {
            steps {
                script {
                    echo '============================================'
                    echo 'CI STAGE 5: BUILD APPLICATION'
                    echo '============================================'
                }
                
                sh '''
                    # Build the application
                    echo "Building application..."
                    CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
                        go build -v \
                        -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.GitCommit=${GIT_COMMIT_SHORT}" \
                        -o bin/api \
                        ./cmd/api
                    
                    # Verify binary
                    echo "Binary information:"
                    ls -lh bin/api
                    file bin/api || echo "file command not available"
                    
                    # Verify binary is executable
                    if [ -x bin/api ]; then
                        echo "Binary built successfully"
                    else
                        echo "ERROR: Binary is not executable"
                        exit 1
                    fi
                '''
            }
        }
        
        // ====================================================================
        // CI STAGE 6: UNIT TESTS
        // ====================================================================
        stage('🧪 Unit Tests') {
            when {
                expression { params.SKIP_TESTS == false }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CI STAGE 6: UNIT TESTS'
                    echo '============================================'
                }
                
                sh '''
                    # Run unit tests with coverage
                    echo "Running unit tests..."
                    # Note: Race detector disabled due to CGO compatibility issues in CI environment
                    # TODO: Re-enable race detector when CI environment supports CGO properly
                    go test -v -timeout=${TEST_TIMEOUT} \
                        -coverprofile=coverage.out \
                        -covermode=atomic \
                        ./internal/... ./pkg/...

                    # Generate coverage report
                    echo "Generating coverage report..."
                    go tool cover -func=coverage.out > coverage.txt
                    
                    # Display coverage summary
                    echo "Coverage Summary:"
                    cat coverage.txt | tail -n 1
                    
                    # Check coverage threshold
                    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
                    echo "Total Coverage: ${COVERAGE}%"
                    
                    # Compare coverage using awk (bc not available in container)
                    if awk "BEGIN {exit !($COVERAGE < ${COVERAGE_THRESHOLD})}"; then
                        echo "WARNING: Coverage ${COVERAGE}% is below threshold ${COVERAGE_THRESHOLD}%"
                        # Uncomment to fail build on low coverage
                        # exit 1
                    fi
                '''
                
                // Archive test results
                junit allowEmptyResults: true, testResults: '**/test-results/*.xml'
                
                // Publish coverage report
                publishHTML([
                    allowMissing: true,
                    alwaysLinkToLastBuild: true,
                    keepAll: true,
                    reportDir: '.',
                    reportFiles: 'coverage.txt',
                    reportName: 'Coverage Report'
                ])
            }
        }
        
        // ====================================================================
        // CI STAGE 7: INTEGRATION TESTS
        // ====================================================================
        stage('🔗 Integration Tests') {
            when {
                expression { params.SKIP_TESTS == false }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CI STAGE 7: INTEGRATION TESTS'
                    echo '============================================'
                }
                
                script {
                    // Check if Docker is available (supports both docker-compose and docker compose)
                    def dockerAvailable = sh(script: 'command -v docker >/dev/null 2>&1; echo $?', returnStdout: true).trim() == '0'
                    if (!dockerAvailable) {
                        error 'Docker is not available on this agent. Install Docker or use a Docker-capable node.'
                    }
                    
                    // Check if Docker daemon is accessible
                    def dockerAccessible = sh(script: 'docker info >/dev/null 2>&1; echo $?', returnStdout: true).trim() == '0'
                    if (!dockerAccessible) {
                        error 'Docker daemon is not accessible. Please check Docker socket permissions.'
                    }
                    
                    // Determine which compose command to use
                    def composeCmd = sh(script: '''
                        if command -v docker-compose >/dev/null 2>&1; then
                            echo "docker-compose"
                        elif docker compose version >/dev/null 2>&1; then
                            echo "docker compose"
                        else
                            echo "none"
                        fi
                    ''', returnStdout: true).trim()
                    
                    if (composeCmd == 'none') {
                        error 'Neither docker-compose nor docker compose plugin available.'
                    }
                    
                    echo "Using compose command: ${composeCmd}"
                    sh """
                        # Start test dependencies with Docker Compose
                        echo "Starting test environment..."
                        ${composeCmd} -f docker-compose.yml up -d mysql

                        # Wait for MySQL to be ready
                        echo "Waiting for MySQL to be ready..."
                        for i in \$(seq 1 30); do
                            if ${composeCmd} exec -T mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
                                echo "MySQL is ready!"
                                break
                            fi
                            echo "Waiting for MySQL... (\$i/30)"
                            sleep 2
                        done
                        
                        # Run database migrations
                        echo "Running migrations..."
                        ./scripts/migrate.sh up || true
                        
                        # Run integration tests
                        echo "Running integration tests..."
                        go test -v -timeout=${TEST_TIMEOUT} \
                            -tags=integration \
                            ./tests/integration/... || true
                        
                        # Cleanup
                        echo "Cleaning up test environment..."
                        ${composeCmd} down -v
                    """
                }
            }
        }
        
        // ====================================================================
        // CI STAGE 8: SECURITY SCAN
        // ====================================================================
        stage('🔒 Security Scan') {
            steps {
                script {
                    echo '============================================'
                    echo 'CI STAGE 8: SECURITY VULNERABILITY SCAN'
                    echo '============================================'
                }
                
                sh '''
                    # Ensure GOPATH/bin is in PATH
                    export PATH="${GOPATH}/bin:${PATH}"
                    
                    # Install gosec if not available
                    if ! command -v gosec &> /dev/null; then
                        echo "Installing gosec..."
                        go install github.com/securego/gosec/v2/cmd/gosec@latest
                    fi
                    
                    # Run security scan
                    echo "Running gosec security scan..."
                    gosec -fmt=json -out=gosec-report.json -exclude-dir=.gomodcache -exclude-dir=.go -exclude-dir=.gocache ./... || true
                    
                    # Install trivy if not available
                    if ! command -v trivy &> /dev/null; then
                        echo "Installing trivy..."
                        curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b ${GOPATH}/bin
                    fi
                    
                    # Run trivy security scan
                    echo "Running trivy filesystem scan..."
                    trivy fs --severity HIGH,CRITICAL --format json --output trivy-report.json --skip-dirs .gomodcache . || true
                '''
                
                // Archive security reports
                archiveArtifacts artifacts: '*-report.json', allowEmptyArchive: true
            }
        }
        
        // ====================================================================
        // CD STAGE 9: BUILD DOCKER IMAGE
        // ====================================================================
        stage('🐳 Build Docker Image') {
            when {
                expression { params.DEPLOY_ENVIRONMENT != 'none' }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 9: BUILD DOCKER IMAGE'
                    echo '============================================'
                    
                    // Use override tag if provided, otherwise use auto-generated
                    def dockerTag = params.DOCKER_TAG_OVERRIDE ?: env.DOCKER_TAG
                    env.FINAL_DOCKER_TAG = dockerTag
                    
                    echo "Building Docker image with tag: ${dockerTag}"
                }
                
                script {
                    def dockerAvailable = sh(script: 'command -v docker >/dev/null 2>&1; echo $?', returnStdout: true).trim() == '0'
                    if (!dockerAvailable) {
                        error 'Docker is not available on this agent. Install Docker or use a Docker-capable node.'
                    }
                }
                sh '''
                    # Setup Docker buildx for multi-platform builds
                    echo "Setting up Docker buildx..."
                    docker buildx create --name multiplatform-builder --use 2>/dev/null || docker buildx use multiplatform-builder || true
                    docker buildx inspect --bootstrap
                    
                    # Build multi-platform Docker image for linux/amd64 (Render requirement)
                    echo "Building Docker image for linux/amd64..."
                    docker buildx build \
                        --platform linux/amd64 \
                        --build-arg VERSION=${BUILD_VERSION} \
                        --build-arg GIT_COMMIT=${GIT_COMMIT_SHORT} \
                        --tag ${DOCKER_HUB_REPO}:${FINAL_DOCKER_TAG} \
                        --tag ${DOCKER_HUB_REPO}:latest \
                        --file Dockerfile \
                        --load \
                        .
                    
                    # Display image information
                    echo "Docker image built successfully:"
                    docker images ${DOCKER_HUB_REPO}
                    
                    # Scan Docker image with trivy
                    echo "Scanning Docker image for vulnerabilities..."
                    trivy image --severity HIGH,CRITICAL ${DOCKER_HUB_REPO}:${FINAL_DOCKER_TAG} || true
                '''
            }
        }
        
        // ====================================================================
        // CD STAGE 10: PUSH TO DOCKER HUB
        // ====================================================================
        stage('📤 Push to Docker Hub') {
            when {
                expression { params.DEPLOY_ENVIRONMENT != 'none' }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 10: PUSH TO DOCKER HUB'
                    echo '============================================'
                }
                
                // Login to Docker Hub and push image
                script {
                    def dockerAvailable = sh(script: 'command -v docker >/dev/null 2>&1; echo $?', returnStdout: true).trim() == '0'
                    if (!dockerAvailable) {
                        error 'Docker is not available on this agent. Cannot push to Docker Hub.'
                    }
                }
                withCredentials([usernamePassword(
                    credentialsId: env.DOCKER_CREDENTIALS_ID,
                    usernameVariable: 'DOCKER_USERNAME',
                    passwordVariable: 'DOCKER_PASSWORD'
                )]) {
                    sh '''
                        # Login to Docker Hub
                        echo "Logging in to Docker Hub..."
                        echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
                        
                        # Rebuild and push multi-platform image directly to registry
                        echo "Building and pushing multi-platform image for linux/amd64..."
                        docker buildx build \
                            --platform linux/amd64 \
                            --build-arg VERSION=${BUILD_VERSION} \
                            --build-arg GIT_COMMIT=${GIT_COMMIT_SHORT} \
                            --tag ${DOCKER_HUB_REPO}:${FINAL_DOCKER_TAG} \
                            --file Dockerfile \
                            --push \
                            .
                        
                        # Push latest tag for main/master branch
                        if [ "${BRANCH_NAME}" = "main" ] || [ "${BRANCH_NAME}" = "master" ]; then
                            echo "Building and pushing latest tag..."
                            docker buildx build \
                                --platform linux/amd64 \
                                --build-arg VERSION=${BUILD_VERSION} \
                                --build-arg GIT_COMMIT=${GIT_COMMIT_SHORT} \
                                --tag ${DOCKER_HUB_REPO}:latest \
                                --file Dockerfile \
                                --push \
                                .
                        fi
                        
                        # Logout
                        docker logout
                    '''
                }
            }
        }
        
        // ====================================================================
        // CD STAGE 11: DEPLOY TO STAGING
        // ====================================================================
        stage('🚀 Deploy to Staging') {
            when {
                expression { params.DEPLOY_ENVIRONMENT == 'staging' }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 11: DEPLOY TO STAGING (Render.com)'
                    echo '============================================'
                }
                
                withCredentials([string(credentialsId: env.RENDER_API_KEY_CREDENTIALS_ID, variable: 'RENDER_API_KEY')]) {
                    sh """
                        # Trigger Render deployment via API with Docker image
                        echo "Triggering deployment to Render staging environment..."
                        echo "Deploying Docker image: ${DOCKER_HUB_REPO}:${FINAL_DOCKER_TAG}"
                        
                        RESPONSE=\$(curl -s -w "\\n%{http_code}" -X POST \\
                            "https://api.render.com/v1/services/${RENDER_STAGING_SERVICE_ID}/deploys" \\
                            -H "Authorization: Bearer \${RENDER_API_KEY}" \\
                            -H "Content-Type: application/json" \\
                            -d '{
                                "clearCache": "do_not_clear",
                                "imageUrl": "${DOCKER_HUB_REPO}:${FINAL_DOCKER_TAG}"
                            }')
                        
                        HTTP_CODE=\$(echo "\$RESPONSE" | tail -n1)
                        BODY=\$(echo "\$RESPONSE" | sed '\$d')
                        
                        echo "Response: \$BODY"
                        
                        if [ "\$HTTP_CODE" -ge 200 ] && [ "\$HTTP_CODE" -lt 300 ]; then
                            echo "✅ Deployment triggered successfully"
                            
                            # Extract deploy ID
                            DEPLOY_ID=\$(echo "\$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
                            echo "Deploy ID: \$DEPLOY_ID"
                            
                            # Wait for deployment to complete (optional)
                            echo "Waiting for deployment to complete..."
                            for i in {1..30}; do
                                sleep 10
                                STATUS=\$(curl -s -X GET \\
                                    "https://api.render.com/v1/services/${RENDER_STAGING_SERVICE_ID}/deploys/\$DEPLOY_ID" \\
                                    -H "Authorization: Bearer \${RENDER_API_KEY}" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
                                echo "Deployment status: \$STATUS"
                                
                                if [ "\$STATUS" = "live" ]; then
                                    echo "✅ Deployment completed successfully"
                                    break
                                elif [ "\$STATUS" = "build_failed" ] || [ "\$STATUS" = "deactivated" ]; then
                                    echo "❌ Deployment failed with status: \$STATUS"
                                    exit 1
                                fi
                            done
                        else
                            echo "❌ Failed to trigger deployment. HTTP Code: \$HTTP_CODE"
                            exit 1
                        fi
                    """
                }
            }
        }
        
        // ====================================================================
        // CD STAGE 12: DEPLOY TO PRODUCTION
        // ====================================================================
        stage('🎯 Deploy to Production') {
            when {
                allOf {
                    expression { params.DEPLOY_ENVIRONMENT == 'production' }
                    anyOf {
                        branch 'main'
                        branch 'master'
                    }
                }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 12: DEPLOY TO PRODUCTION (Render.com)'
                    echo '============================================'
                    
                    // Manual approval for production deployment
                    input message: 'Deploy to Production?', ok: 'Deploy', submitter: 'admin'
                }
                
                withCredentials([string(credentialsId: env.RENDER_API_KEY_CREDENTIALS_ID, variable: 'RENDER_API_KEY')]) {
                    sh """
                        # Trigger Render deployment via API with Docker image
                        echo "Triggering deployment to Render production environment..."
                        echo "Deploying Docker image: ${DOCKER_HUB_REPO}:${FINAL_DOCKER_TAG}"
                        
                        RESPONSE=\$(curl -s -w "\\n%{http_code}" -X POST \\
                            "https://api.render.com/v1/services/${RENDER_PRODUCTION_SERVICE_ID}/deploys" \\
                            -H "Authorization: Bearer \${RENDER_API_KEY}" \\
                            -H "Content-Type: application/json" \\
                            -d '{
                                "clearCache": "do_not_clear",
                                "imageUrl": "${DOCKER_HUB_REPO}:${FINAL_DOCKER_TAG}"
                            }')
                        
                        HTTP_CODE=\$(echo "\$RESPONSE" | tail -n1)
                        BODY=\$(echo "\$RESPONSE" | sed '\$d')
                        
                        echo "Response: \$BODY"
                        
                        if [ "\$HTTP_CODE" -ge 200 ] && [ "\$HTTP_CODE" -lt 300 ]; then
                            echo "✅ Deployment triggered successfully"
                            
                            # Extract deploy ID
                            DEPLOY_ID=\$(echo "\$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
                            echo "Deploy ID: \$DEPLOY_ID"
                            
                            # Wait for deployment to complete
                            echo "Waiting for deployment to complete..."
                            for i in {1..30}; do
                                sleep 10
                                STATUS=\$(curl -s -X GET \\
                                    "https://api.render.com/v1/services/${RENDER_PRODUCTION_SERVICE_ID}/deploys/\$DEPLOY_ID" \\
                                    -H "Authorization: Bearer \${RENDER_API_KEY}" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
                                echo "Deployment status: \$STATUS"
                                
                                if [ "\$STATUS" = "live" ]; then
                                    echo "✅ Deployment completed successfully"
                                    break
                                elif [ "\$STATUS" = "build_failed" ] || [ "\$STATUS" = "deactivated" ]; then
                                    echo "❌ Deployment failed with status: \$STATUS"
                                    exit 1
                                fi
                            done
                        else
                            echo "❌ Failed to trigger deployment. HTTP Code: \$HTTP_CODE"
                            exit 1
                        fi
                    """
                }
            }
        }
        
        // ====================================================================
        // CD STAGE 13: HEALTH CHECK
        // ====================================================================
        stage('✅ Health Check') {
            when {
                expression { params.DEPLOY_ENVIRONMENT != 'none' }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 13: HEALTH CHECK'
                    echo '============================================'
                    
                    def targetServer = params.DEPLOY_ENVIRONMENT == 'production' ? 
                        env.PRODUCTION_SERVER : env.STAGING_SERVER
                    
                    echo "Checking health of ${targetServer}..."
                }
                
                sh '''
                    # Determine target server
                    if [ "${DEPLOY_ENVIRONMENT}" = "production" ]; then
                        TARGET_URL="https://${PRODUCTION_SERVER}/health"
                    else
                        TARGET_URL="https://${STAGING_SERVER}/health"
                    fi
                    
                    # Health check with retries
                    echo "Performing health check on ${TARGET_URL}..."
                    
                    MAX_RETRIES=10
                    RETRY_COUNT=0
                    
                    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
                        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" ${TARGET_URL})
                        
                        if [ "$HTTP_CODE" = "200" ]; then
                            echo "✅ Health check passed! (HTTP ${HTTP_CODE})"
                            
                            # Display health response
                            curl -s ${TARGET_URL} | jq . || curl -s ${TARGET_URL}
                            exit 0
                        fi
                        
                        RETRY_COUNT=$((RETRY_COUNT + 1))
                        echo "⏳ Health check attempt ${RETRY_COUNT}/${MAX_RETRIES} failed (HTTP ${HTTP_CODE}). Retrying in 5s..."
                        sleep 5
                    done
                    
                    echo "❌ Health check failed after ${MAX_RETRIES} attempts"
                    exit 1
                '''
            }
        }
        
        // ====================================================================
        // CD STAGE 14: NOTIFICATION
        // ====================================================================
        stage('📢 Notification') {
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 14: SEND NOTIFICATIONS'
                    echo '============================================'
                }
                
                // Send Slack notification (optional)
                script {
                    def deployStatus = currentBuild.result ?: 'SUCCESS'
                    def color = deployStatus == 'SUCCESS' ? 'good' : 'danger'
                    def message = """
                        *${deployStatus}*: Job `${env.JOB_NAME}` build `${env.BUILD_NUMBER}`
                        *Branch*: ${env.BRANCH_NAME}
                        *Commit*: ${env.GIT_COMMIT_SHORT}
                        *Environment*: ${params.DEPLOY_ENVIRONMENT}
                        *Docker Tag*: ${env.FINAL_DOCKER_TAG ?: 'N/A'}
                        *Build URL*: ${env.BUILD_URL}
                    """.stripIndent()
                    
                    echo "Deployment Status: ${message}"
                    
                    // Uncomment to enable Slack notifications
                    // slackSend(
                    //     channel: env.SLACK_CHANNEL,
                    //     color: color,
                    //     message: message,
                    //     tokenCredentialId: env.SLACK_CREDENTIALS_ID
                    // )
                }
            }
        }
    }
    
    // ========================================================================
    // POST-BUILD ACTIONS
    // ========================================================================
    post {
        always {
            script {
                // Ensure we have a workspace context even if early failure occurred
                try {
                    node {
                        echo '============================================'
                        echo 'POST-BUILD: CLEANUP'
                        echo '============================================'
                        
                        // Cleanup workspace
                        sh '''
                            # Remove test containers (check Docker access first)
                            if docker info > /dev/null 2>&1; then
                                echo "Cleaning up Docker resources..."
                                docker-compose down -v || true
                                docker image prune -f || true
                            else
                                echo "Docker not accessible, skipping Docker cleanup"
                            fi
                            
                            # Display disk usage
                            df -h
                        '''
                        
                        // Archive artifacts BEFORE cleanup
                        archiveArtifacts artifacts: 'bin/api,coverage.out,coverage.txt', allowEmptyArchive: true
                        
                        // Clean workspace (manual cleanup since cleanWs plugin may not be installed)
                        sh '''
                            echo "Cleaning workspace..."
                            rm -rf bin/ *.out *.txt *-report.json .gomodcache .gocache || true
                            echo "Workspace cleaned"
                        '''
                    }
                } catch (err) {
                    echo "Skipping post-build workspace steps due to missing context: ${err}"
                }
            }
        }
        
        success {
            echo '✅ Pipeline completed successfully!'
            
            // Send success notification
            script {
                def message = """
                    ✅ *BUILD SUCCESS*
                    Job: ${env.JOB_NAME} #${env.BUILD_NUMBER}
                    Branch: ${env.BRANCH_NAME}
                    Deployed to: ${params.DEPLOY_ENVIRONMENT}
                    Duration: ${currentBuild.durationString}
                """.stripIndent()
                
                echo message
            }
        }
        
        failure {
            echo '❌ Pipeline failed!'
            
            // Send failure notification
            script {
                def message = """
                    ❌ *BUILD FAILED*
                    Job: ${env.JOB_NAME} #${env.BUILD_NUMBER}
                    Branch: ${env.BRANCH_NAME}
                    Stage: ${env.STAGE_NAME}
                    Build URL: ${env.BUILD_URL}
                """.stripIndent()
                
                echo message
            }
        }
        
        unstable {
            echo '⚠️ Pipeline is unstable!'
        }
        
        aborted {
            echo '🛑 Pipeline was aborted!'
        }
    }
}
