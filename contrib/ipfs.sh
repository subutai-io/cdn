#!/bin/bash

objects="QmepSBbWspyPkiQFMqkrqKeZpFr913VLrASnVqCGU1gW4c Qme9M21vV6q6GcXTiwWdW2dTFHMxRtsyco5EzJQw3QcNpC QmRQe2U3mzCUJLfg233mJRYDsYyYdJaA9b1wS5vbSXupur QmeKcFb8Bsg1mfn7aWuykduX75JBfBvZJifxbwsr9eGg7h QmetLuhCj5FrF5Xcu2d5FGyxsu4GWtsjGjFSi3FXtR73mY QmeZ916vgARQkTWup4whQnsoMGc74MeB89mghUzcPLhwGM QmS3apSdNsCkmeWmvDfMNoyQErKNZCUjDBUWFoersC75Qg QmZqg836oPxRKaK29q4ZHENWi8TzttYEeWX8qMAvxCH5jk QmYwGCDZVHMrFdUn67cFU5VX8knLXpAPupKTPQzbPMVgxq Qmd3y6iWnra9VZ6RuVBiMsbhahE9fb2sHMKWy89j7vNu6a Qmdxvze7CswmLAnNu2Wh9ptDhtykL8SP5xqQBmzZn76iGB QmQ8TYJCPH3EwL4aro9Hfwccie2RBRkstz34Xf8q2kaqKf QmZpqKxkn7dVB4sdSZKsEEvairPoXFSnBypp3bMzheybSD QmTHXFEYT6TUv2WYxx8VqSBAp6EmUCF8PxB7a3KjHpAKer QmP1y96Ggat3CyXRze8wrrnEwqtbW8THike2WVK1zdXhSy QmXj1QBTgG7f9cTTKkojh6UswR6fRsAbGPE4dEh9uhZDRv QmVZGtkVRe67nU1fy463X4tyWWh1Ph9nXSTPCHbLcr5mdP QmXzW9AMN3cYG7PqGfgW7UUst47QWieN855oYbXGPZJrCg QmVAnE9PXWJZoX7GP1vMq8fg2go5n5xcDC4tG6sdaX8zQb QmedMn7vhM1PUG7EEo99tD9WAqQMUEbyqHjaz33mzXQkwu"

for object in $objects 
do 
    echo "Pinning ${object}"
    ipfs pin add --progress /ipfs/${object}
        if [ $? -ne 0 ]; then
        echo "Sorry, the object ${object} is not pinned"
        #exit 1
        fi
done
