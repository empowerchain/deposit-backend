# Deposit Backend

This is a "temporary" solution for a backend for the deposit app. Eventually, this will be implemented on the EmpowerChain, but we want to test the Deposit App in different scenarios before we get that far. So we are starting with Plastic Credits on the EmpowerChain, and nailing down the specifics of the Deposit Schemes through this repo.

Take a look at the White Paper for more details: https://github.com/empowerchain/empowerchain

## Auth

Auth is documented under [commons/auth.md](commons/auth.md).

## Logical flow

The following illustration documents the logical flow of data in the application:

![Logical flow](logic_flow.jpg)

## Data model

The following illustration documents the current data model and how the relations are built

![Data model](data_model.jpg)