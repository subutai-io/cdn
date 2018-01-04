use strict;
use warnings;
use Test::More;
use URI;
use lib qw(./lib);

use_ok('Gorjun::Build');

ok my $gb = Gorjun::Build->new( 
    remote => URI->new('https://github.com/marcoarthur/gorjun.git'),
    branch => 'dev', 
    commit => '4d52af7d847f1464c4c072a0552d457186474b12'),
'Created a gorjun build';
note($gb->status);

ok $gb->start_gorjun( logs => '/tmp/gorjun.logs' ), 'Started gorjun';

ok $gb->stop(), 'Stopped gorjun';

ok $gb->clean, 'Cleaned environment';

done_testing();
