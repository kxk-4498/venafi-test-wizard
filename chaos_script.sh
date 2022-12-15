#!/bin/bash
cat logo.txt | lolcat

while :
do

        #Taking input
        read -p "Enter command(Enter 'chaos help' incase of assistance): " command


        #Splitting
        IFS=' '
        read -ra arr <<< "$command"

        #Switch case
        case "${arr[1]}" in
        #Command Help
        "help") 
        echo "Allowed commands: "
        echo "-> chaos setup"
        echo "-> chaos deploy app"
        echo "-> chaos deploy issuer -sleep <sleep_amount>"
        echo "-> chaos deploy cert <name> <duration> <renewbefore>"
        echo "-> chaos terminate" 
        echo "-> chaos show cert"
        echo "-> chaos get report" 
        echo "-> chaos create network-chaos"
        echo "-> chaos remove network-chaos"
        echo "-> chaos apply file"
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
        kind create cluster --config new-config.yaml 
        docker pull projects.registry.vmware.com/antrea/antrea-ubuntu:latest
        kind load docker-image projects.registry.vmware.com/antrea/antrea-ubuntu:latest
        kubectl apply -f https://github.com/antrea-io/antrea/releases/download/v1.9.0/antrea.yml
        sleep 90
        echo "############################################################################" 
        echo "Setting up cert-manager ... "
        cd ..
        cd cert-manager/
        make K8S_VERSION=1.25 e2e-setup-certmanager
        sleep 30
        cd ..
        cd Venafi-test-wizard/
        echo "############################################################################" 

        echo "Setting up Ingress Controller ... "
        kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.5.1/deploy/static/provider/cloud/deploy.yaml
        sleep 60
        kubectl -n ingress-nginx --address 0.0.0.0 port-forward svc/ingress-nginx-controller 80 &> output_port_80.log &
        kubectl -n ingress-nginx --address 0.0.0.0 port-forward svc/ingress-nginx-controller 443 &> output_port_443.log &
        
        #echo "Setting up deployment... "
        #kubectl apply -f ./application/basic-deploy.yaml
        #echo "Setting up service... "
        #kubectl apply -f ./application/basic-svc.yaml
        #echo "Setting up ingress... "
        #kubectl apply -f ./application/basic-ingress.yaml
        echo "############################################################################" 
        echo "Setting up environment for chaos issuer ... "
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
        #kubectl create namespace chaos
        kubectl apply -f config/samples/issuer.yaml 
        #-n chaos
        echo "Installed chaos issuer. Environment setup complete :) "
        echo "############################################################################"  
        echo ""
        ;;

        "deploy")
        #Installing the issuer.
        if [ ${arr[2]} == "issuer" ]
        then
                echo "############################################################################" 
                echo "Namesapces available :"
                kubectl get namespaces
                echo "############################################################################" 
                #read -p "What namespace would you like to use? " ns_name
                yq ".spec.Scenarios.sleepDuration |= ${arr[4]}" config/samples/issuer.yaml > temp.yaml  
                rm config/samples/issuer.yaml
                mv temp.yaml config/samples/issuer.yaml
                kubectl apply -f config/samples/issuer.yaml 
                #-n $ns_name
                echo "Issuers running: "
                kubectl get ChaosIssuer 
                #-n $ns_name
                echo ""

        #Installing the certificate.       
        elif [ ${arr[2]} == "cert" ]
        then
            echo "############################################################################" 
            echo "Namesapces available :"
            kubectl get namespaces
            echo "############################################################################" 
            #read -p "What namespace would you like to use? " ns_name 

            echo "Creating certificate ..."
            yq ".metadata.name |= ${arr[3]}" config/samples/cert.yaml > temp.yaml  
            rm config/samples/cert.yaml
            mv temp.yaml config/samples/cert.yaml
            yq ".spec.duration |= ${arr[4]}" config/samples/cert.yaml > temp.yaml  
            rm config/samples/cert.yaml
            mv temp.yaml config/samples/cert.yaml
            yq ".spec.renewBefore |= ${arr[5]}" config/samples/cert.yaml > temp.yaml 
            rm config/samples/cert.yaml
            mv temp.yaml config/samples/cert.yaml
            kubectl apply -f config/samples/cert.yaml 
            #-n $ns_name
            echo ""

        elif [ ${arr[2]} == "app" ]
        then
            read -p "What deployment file would you like to use? " deploy_name
            read -p "What service file would you like to use? " svc_name
            read -p "What ingress file would you like to use? " ingress_name
            echo "Setting up deployment... "
            kubectl apply -f ./application/$deploy_name
            echo "Setting up service... "
            kubectl apply -f ./application/$svc_name
            echo "Setting up ingress... "
            kubectl apply -f ./application/$ingress_name
            echo "############################################################################" 
            echo ""
        fi
        ;;
        
        
        #Show condition.
        "show")
        echo "############################################################################" 
        echo "Namesapces available :"
        kubectl get namespaces
        echo "############################################################################" 
        #read -p "What namespace would you like to use? " ns_name 
        echo "Getting certificates ..."
        kubectl get certificates 
        #-n $ns_name 
        echo ""
        ;;

        #Apply files.
        "apply")
        echo "############################################################################" 
        read -p "Choose your file:" file_name
        echo "Applying your file... "
        kubectl apply -f ./myfiles/$file_name
        echo ""
        ;;

        #Create block condition.
        "create")
        echo "############################################################################" 
        echo "Namesapces available :"
        kubectl get namespaces
        read -p "Choose the namespace you would like to create chaos in : " ns_name
        echo "Deploying Network Chaos ... :)"
        kubectl apply -f policies/deny.yaml -n $ns_name
        echo ""
        ;;

        #Create block condition.
        "remove")
        echo "############################################################################" 
        echo "Namesapces available :"
        kubectl get namespaces
        read -p "Choose the namespace you would like to remove chaos in : " ns_name
        echo "Removing Network Chaos ... :)"
        kubectl -n $ns_name delete networkpolicy chaos-rule
        echo ""
        ;;

        #Get report condition.
        "get")
        echo "############################################################################" 
        echo "Generating report .."
        cat output.log | python3 e2e_script.py
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
        kill -9 $( lsof -i :80 -t )
        kill -9 $( lsof -i :443 -t )
        kind delete cluster --name $cluster_name 
        exit 0
        ;;
        esac

done
