<?
$files = [
    '1425424846-127.0.0.1-47074.sls',
    '1425426666-127.0.0.1-32885.sls',
    '1425427199-127.0.0.1-44752.sls',
    '1425427266-127.0.0.1-46839.sls',
    '1425428194-127.0.0.1-46840.sls',
    '1425428218-127.0.0.1-46843.sls',
    '1425428904-127.0.0.1-46903.sls',
    '1425429261-127.0.0.1-46906.sls',
];

foreach ($files as $name) {
    $file = file($name);

    $new = array();

    for ($j = 0; $j < count($file); $j++) {
        if ($j < 2) {
            $new[] = trim($file[$j]);
            continue;
        }
        $line = explode ("\t", $file[$j]);
        for ($i = 1; $i < count($line); $i++) { 
            $line[$i] = ($line[$i] + rand(0, 100)) . ".00";
        }
        $new[] = implode("\t", $line);
    }

    file_put_contents("./$name", implode("\n", $new));
}

