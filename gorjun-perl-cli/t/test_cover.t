use strict;
use warnings;
use Test::More;
use URI;
use File::chdir;
use File::Basename;
use lib qw(./lib);
use File::Path qw(rmtree);
use constant { 
    SHUTDOWN => 6,
    GORJUN_PATH => '/opt/gorjun',
};

BEGIN {
    $ENV{GORJUN_COVER} = 1;
}

use_ok('Gorjun::Build');

# just build gorjun first of all
ok my $gb = Gorjun::Build->new(
    remote => URI->new('https://github.com/marcoarthur/gorjun.git'),
    branch => 'dev',
    commit => 'HEAD'
  ),
  'Created a gorjun build';
note( $gb->status );

# get all modules used in project
my @COVER = do {

    # get from go list all packages used for gorjun
    my $proj = join '', $gb->_mk_github_path;
    map { chomp; $_ }
      grep { /^github/ } qx/go list -f '{{ join .Imports "\\n" }}' $proj/;
};

# run all tests for each module. Reason for this is limitation of
# go test that only give coverage for one selected package.
# See go test -coverpkg option.
for my $module (@COVER) {

    # set gorjun in test mode
    my $name = basename($module);
    $gb->run_test_mode(
        module => $module,
        file   => "cover_${name}.out",
        mode   => 'count',
    );

    # run all tests in t/test.t
    my $return = do 't/test.t';

    # error encountered: report it and shutdown manually
    if ( $@ ) { 
        warn "*" x 80;
        warn "* Cound't finish all tests to cover module: $name";
        warn "* Test gave: $@";
        warn "*" x 80;

        # Shut down server manually
        `curl -s http://$ENV{GORJUN_HOST}:$ENV{GORJUN_PORT}/kurjun/rest/shutdown`;
    }

    # waits gorjun shutdown properly
    sleep SHUTDOWN;

    # remove all files and database from gorjun
    rmtree(GORJUN_PATH) if -d GORJUN_PATH;
}

# collect coverage stats and create a report
$gb->report_code_coverage( show => 'local' );

# clean gorjun
END { rmtree(GORJUN_PATH) if -d GORJUN_PATH; }

done_testing();
