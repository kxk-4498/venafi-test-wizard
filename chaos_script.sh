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
        echo "chaos deploy issuer -sleep <sleep_amount>"
        echo "chaos deploy cert <name> <duration> <renewbefore>"
        echo "chaos terminate" 
        echo "chaos show cert"
        echo "chaos get report" 
        echo ""
        ;;


        #Installing the barebones(Environment setup).
        "setup") 
        #read -p "Name of your cluster: " cluster_name
        #echo "############################################################################" 
        #kind create cluster --name ${cluster_name}
        #echo "############################################################################" 
        #echo "Installing cert-manager ... "
        #kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml
        echo "############################################################################" 
        echo "Installing chaos issuer ... "
        make generate manifests
        go mod tidy
        make install
        make deploy
        make run &> output.log &
        echo "############################################################################"  
        echo "Installing chaos issuer ... "
        yq '.spec.Scenarios.sleepDuration |= "0"' config/samples/issuer.yaml > temp.yaml  
        rm config/samples/issuer.yaml
        mv temp.yaml config/samples/issuer.yaml
        yq '.spec.Scenarios.Scenario1 |= "False"' config/samples/issuer.yaml > temp.yaml  
        rm config/samples/issuer.yaml
        mv temp.yaml config/samples/issuer.yaml
        yq '.spec.Scenarios.Scenario2 |= "False"' config/samples/issuer.yaml > temp.yaml  
        rm config/samples/issuer.yaml
        mv temp.yaml config/samples/issuer.yaml
        kubectl create namespace chaos
        kubectl apply -f config/samples/issuer.yaml -n chaos
        echo "Installed chaos issuer in the newly created chaos namespace. Environment setup complete :) "
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
                read -p "What namespace would you like to use? " ns_name
                yq ".spec.Scenarios.sleepDuration |= ${arr[4]]}" config/samples/issuer.yaml > temp.yaml  
                rm config/samples/issuer.yaml
                mv temp.yaml config/samples/issuer.yaml
                kubectl apply -f config/samples/issuer.yaml -n $ns_name
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
            yq ".metadata.name |= ${arr[3]]}" config/samples/cert.yaml > temp.yaml  
            rm config/samples/cert.yaml
            mv temp.yaml config/samples/cert.yaml
            yq ".spec.duration |= ${arr[4]]}" config/samples/cert.yaml > temp.yaml  
            rm config/samples/cert.yaml
            mv temp.yaml config/samples/cert.yaml
            yq ".spec.renewBefore |= ${arr[5]]}" config/samples/cert.yaml > temp.yaml 
            rm config/samples/cert.yaml
            mv temp.yaml config/samples/cert.yaml
            kubectl apply -f config/samples/cert.yaml -n $ns_name
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

        #Get report condition.
        "get")
        echo "############################################################################" 
        echo "Generating report .."
        cat output.log | python e2e_script.py
        echo "Result pdf created"
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
        kind delete cluster --name $cluster_name 
        exit 0
        ;;
        esac

done

