pipeline {
    environment {
        GOCACHE = "/var/cache/jenkins/.gocache" 
    }
    agent { label 'jenkins-agent-cloud-caching' }
    stages {
        stage("setup"){
            steps {
                sh 'mkdir -p $GOCACHE'
            }
        }
        stage("make check") {
            agent { docker reuseNode: true, image: 'golang:1.13.8-buster' }
            steps {
               script {
		   sh 'make clean'
                   sh 'make'
                   sh 'make test'
               }
            }
        }
    }
}
