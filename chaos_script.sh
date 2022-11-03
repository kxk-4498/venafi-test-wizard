#!/bin/bash
cat logo.txt

while :
do

        #Taking input
        read -p "Enter command(Enter 'chaos help' incase of assistance): " command


        #Splitting
        IFS=' '
        read -ra arr <<< "$command"

        #Switch case
        case "${arr[1]]}" in
        #Command Help
        "help") 
        echo "Allowed commands: "
        echo "chaos setup"
        echo "chaos deploy issuer"
        echo "chaos deploy cert"
        echo "chaos terminate" 
        echo ""
        ;;


        #Installing the barebones(Environment setup).
        "setup") 
        read -p "Name of your cluster: " cluster_name
        echo "############################################################################" 
        kind create cluster --name ${cluster_name}
        echo "############################################################################" 
        echo "Installing cert-manager ... "
        kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml
        echo "############################################################################" 
        echo "Installing chaos issuer ... "
        make generate manifests
        go mod tidy
        make install
        make deploy
        make run &> output.log &
        echo "############################################################################"  
        echo ""
        ;;


        
        "deploy")
        #Installing the issuer.
        if [ ${arr[2]]} == "issuer" ]
        then
                echo "############################################################################" 
                echo "Namesapces available :"
                kubectl get namespaces
                echo "############################################################################" 
                read -p "Would you like to create a new namespace?(yes/no) " create_new_ns_flag
                if [ $create_new_ns_flag == "yes" ]
                then
                    read -p "Enter the name of your namespace: " ns_name
                    kubectl create namespace ${ns_name}
                elif [ $create_new_ns_flag == "no" ]
                then
                    read -p "What namespace would you like to use? " ns_name  
                fi
                kubectl apply -f config/samples/self-signed-issuer_v1alpha1_chaosissuer.yaml -n $ns_name
                echo "Issuers running: "
                kubectl get ChaosIssuer -n $ns_name
                echo ""

        #Installing the certificate.       
        elif [ ${arr[2]]} == "cert" ]
        then
            echo "############################################################################" 
            echo "Namesapces available :"
            kubectl get namespaces
            echo "############################################################################" 
            read -p "What namespace would you like to use? " ns_name 
            echo "Creating certificate ..."
            kubectl apply -f config/samples/certificate_chaosissuer.yaml -n $ns_name
            echo ""
        fi
        ;;
        
        
        #Show condition.
        "show")
        echo "############################################################################" 
        echo "Namesapces available :"
        kubectl get namespaces
        echo "############################################################################" 
        read -p "What namespace would you like to use? " ns_name 
        echo "Getting certificates ..."
        kubectl get certificates -n $ns_name 
        echo ""
        ;;


        #Terminate condition.
        "terminate")
        echo "############################################################################" 
        echo "Clusters available :"
        kind get clusters
        echo "############################################################################" 
        read -p "Name of your cluster: " cluster_name 
        echo "Terminating process ..." 
        kill -9 $( lsof -i :8080 -t )
        kind delete cluster --name sample-test
        exit 0
        ;;
        esac


done

