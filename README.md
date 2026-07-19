# battleship
Implementation of Battle Ship Game in Go.


// Player 1 loads page 
    // fires join API
    // BE creates Game and push to redis with Player 1 information
    // Gets token and game id

// Now game is available for second player to join

// Player 1 gets ships

// Player 1 places a ship
    // Fires placement API
    // Finishes placement (after 3) 
    // Game is now ready to begin (from player 1 perspective)

// Player 2 loads page
    // fires join API
        // BE searches for active games in ready for join
            // push to redis with Player 2 information into Game object
            // return the game id and token.
