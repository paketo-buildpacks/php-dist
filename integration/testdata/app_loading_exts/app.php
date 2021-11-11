<?php


    foreach (get_loaded_extensions() as $ext) {
        print("LOADED:" . $ext . "\n");
    }

    while (true) {
        sleep(10);
    }
?>
