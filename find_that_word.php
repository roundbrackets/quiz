<?

define('FILE', './word.list'); 
$word_list = file(FILE);

//$word_list = file('./word.list2');
//$word_list = array_slice($word_list, 0, 500);
//$word_list = [ 'ab', 'ac', 'acacabac', 'arac' ];

$size  = [];
$size_index = [];
$alpha_size = [];

for ($i = 0; $i < count($word_list); $i++) {
    $w = $word_list[$i] = trim($word_list[$i]);
    $a = $w{0};

    if (!array_key_exists($a, $alpha_size)) {
        $alpha_size[$a] = [];
    }

    $b = strlen($w);

    if (!isset($alpha_size[$a][$b])) {
        $alpha_size[$a][$b] = [];
    }
    $alpha_size[$a][$b][] = $i;

    $b = strlen($w);
    if (!array_key_exists($b, $size)) {
        $size[$b] = [ $i ];
        $size_index[] = $b;
    } else {
        $size[$b][] = $i;
    }
}

sort($size_index);
foreach ($alpha_size as $letter => $x) {
    ksort($alpha_size[$letter]);
}

$min_length_word = $size_index[0]*2;

$wordlength = 0;
$longword = [];

for ($i = count($size_index)-1; $i >= 0; $i--) {
    $s = $size_index[$i]; 

    if ($s < $min_length_word) {
        echo "$s is too short we're done\n";
        break;
    }

    if ($s < $wordlength) {
        echo "We have a word that's longer $s < $wordlength\n";
        break;
    }

    echo "Starting with size $s/".count($size[$s])."\n";

    foreach ($size[$s] as $sr) {
        $word       = $word_list[$sr];
        $subwords   = find_subwords($word); 

        echo "Processing $word\n";

        foreach ($subwords as $w) {
            if ($word == $w) {
                continue;
            }

            if (follow_paths(substr($word, strlen($w)))) {
                $file = FILE;
            $x = trim(`egrep "^$w$" $file`);
            if (empty($x)) {
                echo __LINE__." No $w $x!!\n";
                exit;
            }
                echo __LINE__." Match for $w\n";
                echo __LINE__." Hurray, $word matches.\n";
                $wordlength = strlen($word);
                $longword[] = $word;

                // We could end here, but in this case it finds all the words 
                // of the same length.
                // egrep -v \ 
                // "anti|acetates|encephalographically|acetate|ally|ethanolamines" \
                // word.list2 > word.list2
                // Will yield a long list of matching 24 letter words.
                break 1;
            }
        }

        if (!in_array($word, $longword)) {
            echo "Bad word $word!\n";
        }
    }
}

if ($wordlength > 0) {
    echo "Found one or more words: ".implode($longword,',').". They are $wordlength chars long.";
} else {
    echo "Found no words.";
}

///////// Functions

function find_subwords($word, $first = false) {
    global $word_list, $alpha_size;

    $len    = strlen($word);
    $first  = $word[0];

    // Find all the subwords for word.
    // $alpha_size = {
    //      'a' => {
    //          5 => { // word 5 chars long, starting with a
    //              3, // index in word_list array
    //              8,
    //              ...
    //          },
    //          22 => { ... },
    //          ...
    //      }
    // }
    //
    // 'block' == $word_list[3] 
    //
    // Size keys in 
    //
    $subwords = [];
    foreach ($alpha_size[$first] as $key => $val) {
        // $key == word length
        // $val == array of index references to words
        // in word_list that start with $first and are $key
        // characters long,

        // $alpha_size[$first] keys are not sorted. They could be.
        if ($key > $len) {
            continue;
        }

        // loop thru each word.
        foreach ($val as $r) {
            $comp_word = $word_list[$r];
            $good = true;

            // Compare letter by letter. The first letter
            // always matches. 
            for($i = 1; $i < strlen($comp_word); $i++) {
                // If they don't match, move on.
                if ($comp_word[$i] != $word[$i]) {
                    $good = false;
                    continue 2;
                }
            }

            // It matched.
            $subwords[] = $comp_word;
        }
    }
    return $subwords;
}

function follow_paths ($word) {
    global $word_list, $min_length_word;

    // There can be no word shorter than the shortest word.
    if (strlen($word) < $min_length_word) {
        return false;
    }

    // Find all possible subwords.
    $subwords   = find_subwords($word); 

    if (empty($subwords)) {
        return false;
    }
            $file = FILE;

    // Word matches one of the subwords, we're done.
    foreach ($subwords as $w) {
        if ($w == $word) {
            $x = trim(`egrep "^$w$" $file`);
            if (empty($x)) {
                echo __LINE__." No $w $x!!\n";
                exit;
            }
            echo __LINE__." Match for $w\n";
            return true;
        }
    }

    // Remove the subword from the beginning of word, then
    // recurse for each possibility.
    foreach ($subwords as $w) {
        $m = substr($word, strlen($w));

        // We exit on first match.
        if (follow_paths($m)) {
            $x = trim(`egrep "^$w$" $file`);
            if (empty($x)) {
                echo __LINE__." No $w $x!!\n";
                exit;
            }
            echo __LINE__." Match for $w\n";
            return true;
        }
    }

    // None of the subwords matched. Word is not a composit.
    return false;
}
