# Replicated-Blockchain
CS270 Project-2: Replicated blockchain

## Details of the project in the PDF [here](https://github.com/AartiJivrajani/Replicated-Blockchain/blob/master/CS271_Project2.pdf)

## Deployment Details

The project requirement was to assume 3 clients.
Feel free to poke around the project and run it too! 


Please follow the below steps for the same(PS: The project assumes that all the clients are running on localhost on different ports)

```bash
cd $GOPATH/Replicated-Blockchain/client
# run the below commands on 3 different terminals
go run main.go --client_id=1
go run main.go --client_id=2
go run main.go --client_id=3
```