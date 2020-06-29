from sdk import *

addr_list = addresses()

_pid_fail = "id_20063"
_proposer = addr_list[0]
_initial_funding = (int("2") * 10 ** 9)
_each_funding = (int("5") * 10 ** 9)
_funding_goal_general = (int("10") * 10 ** 9)


def test_pass_finalize_proposal():
    _prop = Proposal(_pid_fail, "general", "headline", "proposal for vote", _proposer, _initial_funding)

    # create proposal
    _prop.send_create()
    time.sleep(3)
    encoded_pid = _prop.pid

    # 1st fund
    fund_proposal(encoded_pid, _funding_goal_general, addr_list[0])

    # 2nd fund
    # fund_proposal(encoded_pid, _each_funding, addr_list[1])
    # check_proposal_state(encoded_pid, ProposalStateActive, ProposalStatusVoting)

    # 1st vote --> 25%
    vote_proposal(encoded_pid, OPIN_NEGATIVE, url_0, addr_list[0])
    query_proposal(encoded_pid)
    # # 2nd vote --> 25%
    vote_proposal(encoded_pid, OPIN_NEGATIVE, url_1, addr_list[0])
    query_proposal(encoded_pid)

    # 3rd vote --> 50%
    # vote_proposal(encoded_pid, OPIN_POSITIVE, url_2, addr_list[0])
    #
    # # 4th vote --> 75%
    # vote_proposal(encoded_pid, OPIN_NEGATIVE, url_3, addr_list[0])
    # check_proposal_state(encoded_pid, ProposalStatePassed, ProposalStatusCompleted)

    time.sleep(3)


if __name__ == "__main__":
    # test pass a proposal
    test_pass_finalize_proposal()

    print "#### ACTIVE PROPOSALS: ####"
    query_proposals(0x31)

    print "#### FAILED PROPOSALS: ####"
    query_proposals(0x33)

    print "#### FINALIZED PROPOSALS: ####"
    query_proposals(0x34)

    print "#### FINALIZEFAILED PROPOSALS: ####"
    query_proposals(0x35)