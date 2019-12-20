pipeline {

  agent any

  options {
      timeout(time: 5, unit: 'MINUTES')
      timestamps()
  }
  
  environment {
    // Repository-specific variables
    registry = "omniex/quickfix-go" // ONLY change the registry when using this Jenkinsfile for any poms repo that needs to be published

    registryCredential = 'dockerhub'
    dockerImage = ''

    // Build variables
    major = ''
    minor = ''
    patch = ''
    tag = ''

    GITHUB_TOKEN = credentials('9fc1a29f-e49f-4a29-93b3-4a88b045bd2b')
  }

  stages {
    stage('Build') {
      steps {
        script {
          // find the latest tag from remote, default to 1.0.0 if it doesn't exist
          def gitAuth = env.GIT_URL.replaceAll("github.com", "omniex-deployer:${GITHUB_TOKEN}@github.com")
          def command = $/git ls-remote --quiet --tags --refs ${gitAuth} | awk -v def=1.0.0 -F\\\/ '{ print $3 } END { if(NR==0) {print def} }' | sort -V | tail -n 1/$
          def version = sh(returnStdout: true, script: command).trim()

          echo "found version ${version}"

          def versions = version.split('\\.')
          major = versions[0]
          minor = versions[1]
          patch = versions[2]

          dockerImage = docker.build("$registry:latest", "--build-arg GITHUB_TOKEN=$GITHUB_TOKEN .")
        }
      }
    }

    stage('Publish') {
      agent any
      when {
        expression {
          return ("${env.GIT_BRANCH}" == 'master' || "${env.GIT_BRANCH}" == 'release')
        }
      }

      steps {
        script {
          if (env.GIT_BRANCH == 'master') {
            def new_patch = patch as Integer
            new_patch++

            tag = "${major}.${minor}.${new_patch}"
          } else { // release branch
            def new_minor = minor as Integer
            new_minor++

            tag = "${major}.${new_minor}.0"
          }

          // Push to Docker
          docker.withRegistry('', registryCredential) {
            dockerImage.push()
            dockerImage.push(tag)
          }

          // Push to Github
          def gitAuth = env.GIT_URL.replaceAll("github.com", "omniex-deployer:${GITHUB_TOKEN}@github.com")
          sh("""
            git tag -a ${tag} -m \"[Jenkins CI] ${tag}\"
            git push ${gitAuth} ${tag}
          """)
        }
      }
    }

    stage('Clean') {
      steps {
        // Delete the Docker image to preserve disk space on Jenkins
        sh 'docker images -q "${registry}*" | uniq | xargs --no-run-if-empty docker rmi -f'
        // Delete local tags
        sh 'git tag | xargs git tag -d'
      }
    }
  }

  post {
    failure {
      script {
        def subject = "[Jenkins] ${env.JOB_NAME} - Build #${currentBuild.number} - ${currentBuild.currentResult}"
        slackSend color: "#ffcc00", channel: "#eng-poms-jenkins", message: (subject + " (<${env.BUILD_URL}|Open>)")
      }
    }

    fixed {
      script {
        def subject = "[Jenkins] ${env.JOB_NAME} - Build #${currentBuild.number} - ${currentBuild.currentResult}"
        slackSend color: "#339900", channel: "#eng-poms-jenkins", message: (subject + " (<${env.BUILD_URL}|Open>)")
      }
    }
  }
}

